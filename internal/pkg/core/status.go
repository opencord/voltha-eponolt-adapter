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
	"fmt"
	"time"

	"github.com/opencord/voltha-lib-go/v3/pkg/adapters/adapterif"
	"github.com/opencord/voltha-openolt-adapter/internal/pkg/core/l2oam"
)

var startWatchStatus bool = false

// WatchStatus ovserves the setting file is updated or not
func WatchStatus(ctx context.Context, cp adapterif.CoreProxy) {
	if startWatchStatus {
		return
	}
	startWatchStatus = true
	logger.Debug(ctx, "Start WatchStatus()")
	ticker := time.NewTicker(1000 * time.Millisecond)
	for range ticker.C {
		onuList, err := l2oam.ReadOnuStatusList()
		if err == nil {
			for i, onu := range onuList {
				if onu.RebootState == "reboot" {
					logger.Debug(ctx, fmt.Sprintf("WatchStatus() reboot: %s", onu.ID))
					// reboot flag is set to off immediately
					onuList[i].RebootState = ""

					// find onu using MAC address
					onuDevice := FindL2oamDevice(onu.MacAddress)
					if onuDevice == nil {
						logger.Debug(ctx, fmt.Sprintf("WatchStatus() device not found: %s", onu.ID))
						continue
					}

					// send Reset ONU message
					if err := onuDevice.send(l2oam.GenerateSetResetOnu(l2oam.OnuPkgType)); err != nil {
						continue
					}
					_, err := onuDevice.waitResponse(ResponseTimer)
					if err != nil {
						logger.Error(ctx, fmt.Sprintf("[%s] reset ONU Send Error: %v", onuDevice.getDeviceName(), err))
						continue
					}
				}
			}
			if err := l2oam.WriteOnuStatusList(onuList); err != nil {
				continue
			}
		}
	}
}
