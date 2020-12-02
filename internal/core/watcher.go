/*
 * Copyright 2020-present Open Networking Foundation

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
	"github.com/opencord/voltha-lib-go/v4/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sync"
	"time"
)

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
	for {
		logger.Debug(ctx, "fetching-pods")
		services, err := w.client.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=bbsim"})
		if err != nil {
			panic(err.Error())
		}
		logger.Debugf(ctx, "There are %d services in the cluster", len(services.Items))
		w.handleServices(ctx, services)
		time.Sleep(10 * time.Second)
	}
}

func (w *Watcher) handleServices(ctx context.Context, services *v1.ServiceList) {
	// TODO if a service is removed we'll want to remove the related entries
	for _, service := range services.Items {
		if err := w.queryService(ctx, service); err != nil {
			logger.Errorw(ctx, "error-while-reading-from-service", log.Fields{"error": err.Error()})
		}
	}
}

func (w *Watcher) queryService(ctx context.Context, service v1.Service) error {
	endpoint := fmt.Sprintf("%s.%s.svc:%d", service.Name, service.Namespace, w.config.BBsimSadisPort)
	logger.Infow(ctx, "querying-service", log.Fields{"endpoint": endpoint})

	res, err := http.Get(fmt.Sprintf("http://%s/v2/static", endpoint))

	if err != nil {
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

	//for _, entry := range result.Sadis.Entries {
	//	switch entry.(type) {
	//	case SadisOltEntry:
	//	case *SadisOltEntry:
	//		logger.Infow(ctx, "olt-entry", log.Fields{"entry": entry})
	//	case SadisOnuEntryV2:
	//	case *SadisOnuEntryV2:
	//		logger.Infow(ctx, "onu-entry", log.Fields{"entry": entry})
	//	default:
	//		logger.Warnw(ctx, "unknown-entity", log.Fields{"entry": entry})
	//	}
	//}

	for _, entry := range result.Sadis.Entries {
		if entry.HardwareIdentifier != "" {
			e := SadisOltEntry{
				ID:                 entry.ID,
				HardwareIdentifier: entry.HardwareIdentifier,
				IPAddress:          entry.IPAddress,
				NasID:              entry.NasID,
				UplinkPort:         entry.UplinkPort,
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
