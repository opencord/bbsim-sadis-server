/*
 * Copyright 2020-2023 Open Networking Foundation (ONF) and the ONF Contributors

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

package utils

import (
	"flag"
	"fmt"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
)

const (
	defaultLogLevel       = "WARN"
	defaultLogFormat      = "json" // or "console"
	defaultBBsimSadisPort = 50074
)

type ConfigFlags struct {
	LogLevel       string
	LogFormat      string
	Kubeconfig     string
	BBsimSadisPort int
}

func NewConfigFlags() *ConfigFlags {
	flags := &ConfigFlags{
		LogLevel:       defaultLogLevel,
		LogFormat:      defaultLogFormat,
		Kubeconfig:     "",
		BBsimSadisPort: defaultBBsimSadisPort,
	}
	return flags
}

func (cf *ConfigFlags) ParseCommandArguments() {
	help := fmt.Sprintf("Log level (debug, infor, warn, error)")
	flag.StringVar(&(cf.LogLevel), "log_level", defaultLogLevel, help)

	help = fmt.Sprintf("Log format (json or console)	")
	logFormat := flag.String("log_format", defaultLogFormat, help)

	flag.StringVar(&(cf.Kubeconfig), "kubeconfig", "", "Absolute path to the kubeconfig file")
	flag.IntVar(&(cf.BBsimSadisPort), "bbsim_sadis_port", defaultBBsimSadisPort, "The port on which BBSim exposes the Sadis server")

	flag.Parse()

	if *logFormat != log.CONSOLE && *logFormat != log.JSON {
		panic(fmt.Sprintf("log_format is invalid, allowed values are: %s, %s", log.JSON, log.CONSOLE))
	}

	cf.LogFormat = *logFormat
}
