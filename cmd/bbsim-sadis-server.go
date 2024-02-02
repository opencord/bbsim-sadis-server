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

package main

import (
	"context"
	"github.com/opencord/bbsim-sadis-server/internal/core"
	"github.com/opencord/bbsim-sadis-server/internal/utils"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sync"
)

var (
	logger log.CLogger
	cf     *utils.ConfigFlags
)

func init() {
	ctx := context.Background()

	cf = utils.NewConfigFlags()
	cf.ParseCommandArguments()

	// Setup this package so that it's log level can be modified at run time
	var err error
	logger, err = log.RegisterPackage(cf.LogFormat, log.ErrorLevel, log.Fields{})
	if err != nil {
		panic(err)
	}

	// Set the instance ID as the hostname
	instanceID := utils.GetHostName()

	logLevel, err := log.StringToLogLevel(cf.LogLevel)

	if err != nil {
		logger.Errorw(ctx, "provided-log-level-is-not-valid", log.Fields{"err": err, "providedLevel": cf.LogLevel})
	}

	//Setup default logger - applies for packages that do not have specific logger set
	if _, err := log.SetDefaultLogger(cf.LogFormat, logLevel, log.Fields{"instanceId": instanceID}); err != nil {
		logger.With(log.Fields{"error": err}).Fatal(ctx, "Cannot setup logging")
	}

	// Update all loggers (provisioned via init) with a common field
	if err := log.UpdateAllLoggers(log.Fields{"instanceId": instanceID}); err != nil {
		logger.With(log.Fields{"error": err}).Fatal(ctx, "Cannot setup logging")
	}

	// Update all loggers to log level specified as input parameter
	log.SetAllLogLevel(logLevel)

	core.SetupLogger(logLevel, cf.LogFormat)
}

func main() {
	ctx := context.Background()
	logger.Info(ctx, "bbsim-sadis-server-started")

	var config *rest.Config
	var err error

	// if kubeconfig is provided use that, otherwise assume we're running within the cluster
	if cf.Kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", cf.Kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	store := core.NewStore()

	watcher := core.NewWatcher(clientset, store, cf)
	server := core.NewServer(store)

	wg := sync.WaitGroup{}

	wg.Add(2)

	go watcher.Watch(ctx, &wg)
	go server.StartSadisServer(&wg)

	wg.Wait()
}
