/*
Copyright 2019 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/wonderivan/logger"
	"keep/constants"
	"keep/pkg/util"
	"os"
	"strings"
	"time"
)

// NewDefaultEdgeAgentConfig returns a full EdgeCoreConfig object
func NewDefaultEdgeAgentConfig() *EdgeAgentConfig {
	hostnameOverride, _ := os.Hostname()
	localIP, _ := util.GetLocalIP(hostnameOverride)
	logger.Info(localIP)
	return &EdgeAgentConfig{
		DataBase: &DataBase{
			DataSource: DataBaseDataSource,
		},
		Modules: &Modules{
			HealthzAgent: &HealthzAgent{
				Enable:                    false,
				CpuUsage:                  0.0,
				DefaultEdgeHealthInterval: constants.DefaultEdgeHealthInterval,
				Cpu:                       nil,
				Mem:                       nil,
				DiskPartitionStat:         nil,
				DiskIOCountersStat:        nil,
				NetIOCountersStat:         nil,
				DeviceMqttTopics:          strings.Split(constants.DefaultDeviceMqttTopics, ";"),
			},
			LogsAgent: &LogsAgent{
				Enable:      true,
				LogLevel:    6,
				LogTime:     time.Now(),
				LogFileName: "",
				LogInfo:     "",
				LogFiles:    []string{constants.DefaultEdgeLogFiles},
			},
			EdgePublisher: &EdgePublisher{
				Enable:            true,
				HTTPServer:        constants.DefaultHttpServer,
				Port:              constants.DefaultCloudHttpPort,
				ServePort:         constants.DefaultEdgePort,
				Heartbeat:         constants.DefaultEdgeHeartBeat,
				EdgeMsgQueens:     []string{constants.DefaultLogsTopic, constants.DefaultDataTopic},
				TLSCAFile:         "",
				TLSCertFile:       "",
				TLSPrivateKeyFile: "",
			},
			EdgeTwin: &EdgeTwin{
				Enable:         true,
				SqliteFilePath: constants.DefaultEdgeTwinSqliteFilePath,
			},
		},
	}
}

// NewMinEdgeCoreConfig returns a common EdgeCoreConfig object
func NewMinEdgeCoreConfig() *EdgeAgentConfig {
	hostnameOverride, _ := os.Hostname()
	localIP, _ := util.GetLocalIP(hostnameOverride)
	logger.Info(localIP)
	return &EdgeAgentConfig{
		DataBase: &DataBase{
			DataSource: DataBaseDataSource,
		},
		Modules: &Modules{
			HealthzAgent: &HealthzAgent{
				Enable:             true,
				Cpu:                nil,
				Mem:                nil,
				DiskPartitionStat:  nil,
				DiskIOCountersStat: nil,
				NetIOCountersStat:  nil,
			},
		},
	}
}
