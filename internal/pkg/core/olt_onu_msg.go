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

	"github.com/golang/protobuf/ptypes"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	ic "github.com/opencord/voltha-protos/v3/go/inter_container"
	oop "github.com/opencord/voltha-protos/v3/go/openolt"
)

// AdditionalMessage is used for something...
type AdditionalMessage struct {
}

func (dh *DeviceHandler) receivedMsgFromOnu(ctx context.Context, msg *ic.InterAdapterMessage) error {
	msgBody := msg.GetBody()
	ind := &oop.OnuIndication{}

	err := ptypes.UnmarshalAny(msgBody, ind)
	if err != nil {
		logger.Debugw(ctx, "cannot-unmarshal-onu-indication-body", log.Fields{"error": err})
		return err
	}

	var info AdditionalMessage
	err = json.Unmarshal(ind.SerialNumber.VendorSpecific, &info)
	if err != nil {
		logger.Debugw(ctx, "cannot-unmarshal-additional-message", log.Fields{"error": err})
		return err
	}

	onuDevice := FindL2oamDeviceByDeviceID(msg.Header.ToDeviceId)
	if onuDevice == nil {
		logger.Error(ctx, "receivedMsgFromOnu() OnuDevice is not found. deviceId=%s", msg.Header.ToDeviceId)
		return nil
	}

	id := msg.Header.Id
	switch id {
	case "reboot":
		onuDevice.reboot(ctx)
	case "disable":
		logger.Warn(ctx, "unsupported message, Id=%s", id)
	case "reenable":
		logger.Warn(ctx, "unsupported message, Id=%s", id)
	}
	return nil
}

/*
func (dh *DeviceHandler) sendMsgToOnu(ctx context.Context, deviceID string, topic string, msg string, option interface{}) error {
	bytes, _ := json.Marshal(option)
	info := oop.OnuIndication{
		SerialNumber: &oop.SerialNumber{
			VendorSpecific: bytes,
		},
	}

	err := dh.AdapterProxy.SendInterAdapterMessage(ctx, &info,
		ic.InterAdapterMessageType_ONU_IND_REQUEST,
		dh.device.Type, topic,
		deviceID, dh.device.Id, msg)

	if err != nil {
		logger.Debugw(ctx, "err-sending-message", log.Fields{"device-id": deviceID, "topic": topic, "error": err})
	}
	return err
}
*/
