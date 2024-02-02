/*
 * Copyright 2020-2024 Open Networking Foundation (ONF) and the ONF Contributors

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opencord/bbsim-sadis-server/internal/utils"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sync"
	"time"
)

const attemptLimit = 10

type Watcher struct {
	client *kubernetes.Clientset
	store  *Store
	config *utils.ConfigFlags
}

func NewWatcher(client *kubernetes.Clientset, store *Store, cf *utils.ConfigFlags) *Watcher {
	return &Watcher{
		client: client,
		store:  store,
		config: cf,
	}
}

func (w *Watcher) Watch(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// we need to watch for PODs, services can't respond to requests if the backend is not there
	// note that when this container starts we receive notifications for all of the existing pods

	watcher, err := w.client.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{LabelSelector: "app=bbsim"})
	if err != nil {
		logger.Fatalw(ctx, "error-while-watching-pods", log.Fields{"err": err})
	}

	ch := watcher.ResultChan()
	for event := range ch {

		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			logger.Fatalw(ctx, "unexpected-type-while-watching-pod", log.Fields{"object": event.Object})
		}

		logger.Debugw(ctx, "received-pod-event", log.Fields{"object": event.Type, "pod": pod.Name})
		if event.Type == watch.Deleted {
			// TODO remove sadis entries
			logger.Debug(ctx, "pod-has-been-removed")
		}

		if event.Type == watch.Added || event.Type == watch.Modified {
			// fetch the sadis information and store them

			// the pod is ready only if all the containers in it are ready,
			// for now the BBSim pod only has 1 container, but things may change in the future, so keep the loop
			ready := true

			if len(pod.Status.ContainerStatuses) == 0 {
				// if there are no containers in the pod, then it's not ready
				ready = false
			}

			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					// if one of the container is not ready, then the entire pod is not ready
					ready = false
				}
			}

			logger.Debugw(ctx, "received-event-for-bbsim-pod", log.Fields{"pod": pod.Name, "namespace": pod.Namespace,
				"release": pod.Labels["release"], "ready": ready, "podIp": pod.Status.PodIP})

			// as soon as the pod is ready cache the sadis entries
			if ready {
				if err := w.queryPod(ctx, pod.Status.PodIP, 0); err != nil {
					logger.Errorw(ctx, "failed-to-load-sadis-config-from-bbsim",
						log.Fields{"pod": pod.Name, "namespace": pod.Namespace, "release": pod.Labels["release"], "err": err})
				}
			}
		}

	}
}

func (w *Watcher) queryPod(ctx context.Context, ip string, attempt int) error {
	endpoint := fmt.Sprintf("%s:%d", ip, w.config.BBsimSadisPort)
	logger.Infow(ctx, "querying-service", log.Fields{"endpoint": endpoint})

	client := http.Client{Timeout: 5 * time.Second}

	res, err := client.Get(fmt.Sprintf("http://%s/v2/static", endpoint))

	if err != nil {
		if attempt < attemptLimit {
			logger.Warnw(ctx, "error-while-reading-from-service-retrying", log.Fields{"error": err.Error()})
			// if there is an error and we have attempt left just retry later
			time.Sleep(1 * time.Second)
			return w.queryPod(ctx, ip, attempt+1)
		}

		return err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	var result SadisConfig

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&result); err != nil {
		logger.Errorw(ctx, "cannot-decode-sadis-response", log.Fields{"error": err.Error()})
		return err
	}

	logger.Debugw(ctx, "fetched-sadis-config", log.Fields{
		"endpoint":          endpoint,
		"entries":           len(result.Sadis.Entries),
		"bandwidthProfiles": len(result.BandwidthProfile.Entries),
	})

	for _, entry := range result.Sadis.Entries {
		if entry.HardwareIdentifier != "" {
			e := SadisOltEntry{
				ID:                 entry.ID,
				HardwareIdentifier: entry.HardwareIdentifier,
				IPAddress:          entry.IPAddress,
				NasID:              entry.NasID,
				UplinkPort:         entry.UplinkPort,
				NniDhcpTrapVid:     entry.NniDhcpTrapVid,
			}
			w.store.addOlt(ctx, e)
			continue
		}
		if len(entry.UniTagList) != 0 {
			e := SadisOnuEntryV2{
				ID:         entry.ID,
				NasPortID:  entry.NasPortID,
				CircuitID:  entry.CircuitID,
				RemoteID:   entry.RemoteID,
				UniTagList: entry.UniTagList,
			}
			w.store.addOnu(ctx, e)
			continue
		}
		logger.Warnw(ctx, "unknown-entity", log.Fields{"entry": entry})
	}

	for _, bp := range result.BandwidthProfile.Entries {
		w.store.addBp(ctx, *bp)
	}

	logger.Infow(ctx, "stored-sadis-config", log.Fields{"endpoint": endpoint})

	return nil
}
