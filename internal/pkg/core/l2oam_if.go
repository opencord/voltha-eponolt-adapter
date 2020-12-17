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
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/opencord/voltha-lib-go/v3/pkg/db/kvstore"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-openolt-adapter/internal/pkg/core/l2oam"
	oop "github.com/opencord/voltha-protos/v3/go/openolt"
	"github.com/opencord/voltha-protos/v3/go/voltha"
)

func readTechProfile() {
	address := "172.17.0.1:2379"
	timeout := time.Second * 10

	client, err := kvstore.NewEtcdClient(context.Background(), address, timeout, log.FatalLevel)
	if err != nil {
		return
	}
	defer client.Close(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resp, err := client.Get(ctx, "service/voltha/technology_profiles/EPON/64")
	cancel()
	if resp == nil || err != nil {
		l2oam.OnuPkgType = l2oam.OnuPkgTypeB
		logger.Error(ctx, fmt.Sprintf("readTechProfile() etcd get error. resp=%v, err=%v", resp, err))
		return
	}
	bytes := resp.Value.([]byte)
	bytesString := string(bytes)
	logger.Debug(ctx, fmt.Sprintf("readTechProfile() bytes string=%s", bytesString))
	if strings.Index(bytesString, "\"profile_type\":\"EPON\"") > 0 && strings.Index(bytesString, "\"package_type\":\"B\"") > 0 {
		l2oam.OnuPkgType = l2oam.OnuPkgTypeB
	} else {
		l2oam.OnuPkgType = l2oam.OnuPkgTypeA
	}
	logger.Debug(ctx, fmt.Sprintf("readTechProfile() onu package=%s", l2oam.OnuPkgType))

}

// L2oamGetDeviceInfo provides GetDeviceInfo for L2OAM
func L2oamGetDeviceInfo(ctx context.Context, dh *DeviceHandler) (*oop.DeviceInfo, error) {
	logger.Debug(ctx, "GetDeviceInfo() Start.")
	readTechProfile()

	dstMac := dh.device.MacAddress
	deviceInfo := &oop.DeviceInfo{
		Vendor: "Tibit",
		// Model				string	`protobuf:"bytes,2,opt,name=model,proto3" json:"model,omitempty"`
		// HardwareVersion		string	`protobuf:"bytes,3,opt,name=hardware_version,json=hardwareVersion,proto3" json:"hardware_version,omitempty"`
		// FirmwareVersion		string	`protobuf:"bytes,4,opt,name=firmware_version,json=firmwareVersion,proto3" json:"firmware_version,omitempty"`
		DeviceId: dstMac,
		// DeviceSerialNumber	string	`protobuf:"bytes,17,opt,name=device_serial_number,json=deviceSerialNumber,proto3" json:"device_serial_number,omitempty"`
		PonPorts:       1,
		Technology:     "EPON",
		OnuIdStart:     1,
		OnuIdEnd:       128,
		AllocIdStart:   5121,
		AllocIdEnd:     9216,
		GemportIdStart: 1,
		GemportIdEnd:   65535,
		FlowIdStart:    1,
		FlowIdEnd:      16383,
		// Ranges				[]*DeviceInfo_DeviceResourceRanges	protobuf:"bytes,15,rep,name=ranges,proto3" json:"ranges,omitempty"`
	}
	olt := FindL2oamDevice(dstMac)
	if olt == nil {
		olt, _ = NewL2oamOltDevice(dstMac, dh)
		olt.startKeepAlive()
	}

	err00 := sendDiscoverySOLICIT(ctx, olt)
	logger.Debug(ctx, fmt.Sprintf("Sequence returts: %v", err00))
	vendorName, err01 := sendGetVendorName(ctx, olt)
	moduleNumber, err02 := sendGetModuleNumber(ctx, olt)
	manuFacturer, err03 := sendManuFacturerInfo(ctx, olt)
	firmwareVersion, err04 := sendGetFirmwareVersion(ctx, olt)
	macAddress, err05 := sendGetMacAddress(ctx, olt)
	serialNumber, err06 := sendGetSerialNumber(ctx, olt)
	deviceInfo.Vendor = vendorName
	deviceInfo.Model = moduleNumber
	deviceInfo.HardwareVersion = manuFacturer
	deviceInfo.FirmwareVersion = firmwareVersion
	deviceInfo.DeviceId = macAddress
	deviceInfo.DeviceSerialNumber = serialNumber
	logger.Debug(ctx, fmt.Sprintf("Sequence returts: deviceInfo=%v, %v, %v, %v, %v, %v, %v, %v", deviceInfo, err00, err01, err02, err03, err04, err05, err06))

	dh.device.SerialNumber = deviceInfo.DeviceSerialNumber
	olt.setActiveState(ctx, true)
	return deviceInfo, nil
}

// L2oamEnableIndication runs EnableIndication sequence
func L2oamEnableIndication(ctx context.Context, dh *DeviceHandler) {
	logger.Debug(ctx, "L2oamEnableIndication() Start.")

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamEnableIndication() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
		return
	}

	err01 := sendSetHbtxPeriod(ctx, olt)
	err02 := sendSetHbtxTemplate(ctx, olt)
	ponMode, err03 := sendGetPonMode(ctx, olt)
	err04 := sendSetMPCPSync(ctx, olt)
	err05 := sendSetAdminState(ctx, olt, true)
	ponPortAction := sendSetGenericActionCreate(ctx, olt, l2oam.ActionTypeProtocolFilter)
	if ponPortAction == nil {
		return
	}
	actionIDPonPort := binary.BigEndian.Uint32(ponPortAction.GetTrafficProfile())
	err07 := sendSetIngressPort(ctx, olt, actionIDPonPort, true)
	err08 := sendSetCaptureProtocols(ctx, olt, actionIDPonPort)
	ethPortAction := sendSetGenericActionCreate(ctx, olt, l2oam.ActionTypeProtocolFilter)
	if ethPortAction == nil {
		return
	}
	actionIDEthPort := binary.BigEndian.Uint32(ethPortAction.GetTrafficProfile())
	err09 := sendSetIngressPort(ctx, olt, actionIDEthPort, false)
	err10 := sendSetCaptureProtocols(ctx, olt, actionIDEthPort)

	oltDev := olt.(*L2oamOltDevice)
	oltDev.Base.ActionIds = make(map[int]uint32)
	oltDev.Base.ActionIds[ActionIDFilterPonPort] = actionIDPonPort
	oltDev.Base.ActionIds[ActionIDFilterEthPort] = actionIDEthPort
	olt.updateMap()

	err06 := sendSetAdminState(ctx, olt, false)
	olt.setAutonomousFlag(true)

	logger.Debug(ctx, fmt.Sprintf("Sequence returns: ponMode=%s, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v",
		ponMode, err01, err02, err03, err04, err05, err06, err07, err08, err09, err10))

}

// L2oamAfterKeepAlive ... [obsoluted]
func L2oamAfterKeepAlive(ctx context.Context, dh *DeviceHandler) error {



	return nil
}

//var flowVersion int = 1 // run add-flow each OLT
//var flowVersion int = 2 // run add-flow each ONU
var flowVersion int = 3 // run add-flow at specified ONU

// L2oamAddFlow runs add-flow sequence
func L2oamAddFlow(ctx context.Context, dh *DeviceHandler, cmd *L2oamCmd) {
	logger.Debug(ctx, "L2oamAddFlow() Start.")
	if cmd.OnuDeviceID == "" {
		flowVersion = 2
	} else {
		flowVersion = 3
	}

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamAddFlow() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
		return
	}

	if flowVersion == 1 {
		oc := &l2oam.TomiObjectContext{
			Branch:   0x0c,
			Type:     0x0011,
			Length:   4,
			Instance: 0x00000002,
		}
		getTrafficControl := sendGetTrafficControlReferenceTable(ctx, olt, oc)
		if getTrafficControl == nil {
			logger.Error(ctx, "L2oamAddFlow() sendGetTrafficControlReferenceTable() error. ")
			return
		}
		downProfile := sendSetGenericActionCreate(ctx, olt, l2oam.ActionTypeTrafficProfile)
		if downProfile == nil {
			logger.Error(ctx, "L2oamAddFlow() sendSetGenericActionCreate(down) error. ")
			return
		}
		upProfile := sendSetGenericActionCreate(ctx, olt, l2oam.ActionTypeTrafficProfile)
		if upProfile == nil {
			logger.Error(ctx, "L2oamAddFlow() sendSetGenericActionCreate(up) error. ")
			return
		}
		olt.setReferenceTable(getTrafficControl)
		olt.setProfile(downProfile, upProfile)
		olt.updateMap()

		err01 := sendSetTrafficControl(ctx, olt, getTrafficControl.GetReferenceControlDown(), downProfile.GetTrafficProfile())
		err02 := sendSetTrafficControl(ctx, olt, getTrafficControl.GetReferenceControlUp(), upProfile.GetTrafficProfile())

		err03 := sendSetPriority(ctx, olt, downProfile.GetTrafficProfile())
		err04 := sendSetGuranteedRate(ctx, olt, cmd.Cir, downProfile.GetTrafficProfile())
		err05 := sendSetGuranteedRate(ctx, olt, cmd.Cir, upProfile.GetTrafficProfile())
		err06 := sendSetBestEffortRate(ctx, olt, cmd.Pir, downProfile.GetTrafficProfile())
		err07 := sendSetBestEffortRate(ctx, olt, cmd.Pir, upProfile.GetTrafficProfile())

		logger.Debug(ctx, fmt.Sprintf("Sequence results: %v, %v, %v, %v, %v, %v, %v", err01, err02, err03, err04, err05, err06, err07))
	} else if flowVersion == 2 {
		// add-flow to all ONUs
		// Traffic Profile(priority, guranteed rate, best effort rate) are set as common settings
		onuDevices, err := dh.coreProxy.GetChildDevices(ctx, dh.device.Id)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("catnnot-get-child-devices: %v", err))
			return
		}

		isFirst := true
		var downProfile *l2oam.SetGenericActionCreateRes
		var upProfile *l2oam.SetGenericActionCreateRes
		for _, onuDevice := range onuDevices.Items {
			//onu := FindL2oamDevice(onuDevice.MacAddress)
			onu := FindL2oamDeviceByDeviceID(onuDevice.Id)
			if onu == nil {
				logger.Error(ctx, fmt.Sprintf("catnnot-get-onu-device. device-id:%s, mac-address:%s",
					onuDevice.Id, onuDevice.MacAddress))
				continue
			}
			if !(onu.(*L2oamOnuDevice)).Base.isActive() {
				logger.Debug(ctx, fmt.Sprintf("onu is not active. device-id:%s, mac-address:%s",
					onuDevice.Id, onuDevice.MacAddress))
				continue
			}
			logger.Debug(ctx, fmt.Sprintf("add-flow-for-onu. device-id:%s, mac-address:%s",
				onuDevice.Id, onuDevice.MacAddress))

			oc := onu.getObjectContext()
			if oc == nil {
				logger.Error(ctx, fmt.Sprintf("catnnot-get-onu-object-context. device-id:%s, mac-address:%s",
					onuDevice.Id, onuDevice.MacAddress))
				continue
			}
			getTrafficControl := sendGetTrafficControlReferenceTable(ctx, olt, oc)
			if getTrafficControl == nil {
				logger.Error(ctx, "L2oamAddFlow() sendGetTrafficControlReferenceTable() error. ")
				continue
			}
			onu.setReferenceTable(getTrafficControl)
			onu.updateMap()

			// create one Traffic Profile at OLT
			if isFirst {
				downProfile = sendSetGenericActionCreate(ctx, olt, l2oam.ActionTypeTrafficProfile)
				if downProfile == nil {
					logger.Error(ctx, "L2oamAddFlow() sendSetGenericActionCreate(down) error. ")
					break
				}
				upProfile = sendSetGenericActionCreate(ctx, olt, l2oam.ActionTypeTrafficProfile)
				if upProfile == nil {
					logger.Error(ctx, "L2oamAddFlow() sendSetGenericActionCreate(up) error. ")
					break
				}
				olt.setProfile(downProfile, upProfile)
				olt.updateMap()
			}

			err01 := sendSetTrafficControl(ctx, olt, getTrafficControl.GetReferenceControlDown(), downProfile.GetTrafficProfile())
			err02 := sendSetTrafficControl(ctx, olt, getTrafficControl.GetReferenceControlUp(), upProfile.GetTrafficProfile())
			logger.Debug(ctx, fmt.Sprintf("traffic-control/traffic-profile results: %v, %v", err01, err02))

			if isFirst {
				err03 := sendSetPriority(ctx, olt, downProfile.GetTrafficProfile())
				err04 := sendSetGuranteedRate(ctx, olt, cmd.Cir, downProfile.GetTrafficProfile())
				err05 := sendSetGuranteedRate(ctx, olt, cmd.Cir, upProfile.GetTrafficProfile())
				err06 := sendSetBestEffortRate(ctx, olt, cmd.Pir, downProfile.GetTrafficProfile())
				err07 := sendSetBestEffortRate(ctx, olt, cmd.Pir, upProfile.GetTrafficProfile())
				isFirst = false
				logger.Debug(ctx, fmt.Sprintf("traffic-profile-settings-results: %v, %v, %v, %v, %v",
					err03, err04, err05, err06, err07))
			}
		}
	} else if flowVersion == 3 {
		if err := olt.addFlow(ctx, cmd); err != nil {
			logger.Errorf(ctx, "failed-to-add-flow: %v", err)
		}
	}
}

// L2oamAddFlowAndMount adds flow to device and start mount sequence
func (d *DeviceHandler) L2oamAddFlowAndMount(ctx context.Context, onuDeviceID string, vid []byte, ivid []byte) {
	l2oamCmd := &L2oamCmd{
		Type:        "add-flow-dev",
		Tpid:        []byte{0x88, 0xa8},
		Vid:         vid,
		Itpid:       []byte{0x81, 0x00},
		Ivid:        ivid,
		OnuDeviceID: onuDeviceID,
	}
	olt := FindL2oamDeviceByDeviceID(d.device.Id)
	if olt == nil {
		logger.Errorf(ctx, "olt not found.")
	}
	if err := olt.updateFlow(ctx, l2oamCmd); err != nil {
		logger.Errorf(ctx, "failed-to-update-flow. %v", err)
	} else {
		onu := FindL2oamDeviceByDeviceID(onuDeviceID)
		if onu == nil {
			logger.Debug(ctx, fmt.Sprintf("L2oamAddFlowAndMount() FindL2oamDevice() onu not found. deviceId=%s", onuDeviceID))
		} else {
			// start ONU mount sequence
			onu.startMountSequence(ctx, l2oam.OnuPkgType, l2oamCmd)
		}
	}
}

// L2oamAddFlowToDeviceAll adds a flow to specified device.
// If some flows are added already, all flows are removed from all devices before adding.
func L2oamAddFlowToDeviceAll(ctx context.Context, dh *DeviceHandler, cmd *L2oamCmd) error {
	logger.Debug(ctx, "L2oamAddFlowToDeviceAll() Start.")

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamAddFlowToDeviceAll() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
		return nil
	}

	oc := olt.getObjectContext()
	var onuList []*L2oamOnuDevice
	onuDeviceID := cmd.OnuDeviceID
	onu := FindL2oamDeviceByDeviceID(onuDeviceID)
	if onu == nil {
		return nil
	}
	onuDev, ok := onu.(*L2oamOnuDevice)
	if !ok {
		return nil
	}
	if oc != nil {
		onuDevices, err := dh.coreProxy.GetChildDevices(ctx, dh.device.Id)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("catnnot-get-child-devices: %v", err))
			return err
		}
		// create flow added onu list
		reflow := false
		for _, onuDevice := range onuDevices.Items {
			//onu := FindL2oamDevice(onuDevice.MacAddress)
			onu := FindL2oamDeviceByDeviceID(onuDevice.Id)
			if onu != nil {
				if onuDev, ok := onu.(*L2oamOnuDevice); ok {
					if onuDev.Base.FlowAdded {
						onuList = append(onuList, onuDev)
						if onuDeviceID == onuDev.Base.DeviceID {
							reflow = true
						}
					}
				}
			}
		}
		if !reflow {
			onuList = append(onuList, onuDev)
		}
		L2oamRemoveFlowFromDevice(ctx, dh, onuDevices)
	} else {
		onuList = append(onuList, onuDev)
	}
	localCmd := *cmd
	for _, onuDev := range onuList {
		localCmd.OnuDeviceID = onuDev.Base.DeviceID
		if err := L2oamAddFlowToDevice(ctx, dh, &localCmd); err != nil {
			continue
		}
	}

	return nil
}

// L2oamAddFlowToDevice runs add-flow-device sequence
func L2oamAddFlowToDevice(ctx context.Context, dh *DeviceHandler, cmd *L2oamCmd) error {
	if flowVersion == 1 {
		oc := L2oamAddFlowToDeviceDS(ctx, dh, cmd)
		if oc == nil {
			return errors.New("failed-to-add-flow-to-device-for-down-stream")
		}
		L2oamAddFlowToDeviceUS(ctx, dh, oc, cmd)
	} else if flowVersion > 1 {
		logger.Debug(ctx, "L2oamAddFlowToDevice() Start.")

		olt := FindL2oamDeviceByDeviceID(dh.device.Id)
		if olt == nil {
			logger.Error(ctx, fmt.Sprintf("L2oamAddFlowToDevice() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
			return nil
		}

		// add flow for downstream
		oc := olt.getObjectContext()
		var err01 error
		var err02 error
		onuDeviceID := cmd.OnuDeviceID
		// "Generic/Action Create, L2 Switching Domain/Action Inlet entry" is created only first
		if oc == nil {
			oc, err01 = sendSetGenericActionCreateForDS(ctx, olt)
			if oc == nil {
				logger.Error(ctx, fmt.Sprintf("error. %v", err01))
				return errors.New("failed-to-add-flow-to-device-for-down-stream")
			}
			// TODO ignore tpid, vid after the second time
			err02 = sendSetL2SwitchingDomainForDS(ctx, olt, oc, cmd.Tpid, cmd.Vid)
			olt.setL2oamCmd(cmd)
			olt.setObjectContext(oc)
			if oltDev, ok := olt.(*L2oamOltDevice); ok {
				oltDev.Base.FlowAdded = true
			}
			olt.updateMap()
		} else {
			cmd = olt.getL2oamCmd()
		}

		onu := FindL2oamDeviceByDeviceID(onuDeviceID)
		if onu == nil {
			logger.Debug(ctx, fmt.Sprintf("L2oamCmdRequest() FindL2oamDevice() onu not found. deviceId=%s", onuDeviceID))
		} else {
			onuID := onu.getObjectContext()
			if onuID == nil {
				logger.Error(ctx, "catnnot-get-onu-object-context.")
			} else {
				err03 := sendSetDefaultOutlet(ctx, olt, oc, onuID)
				err04 := sendSetL2SwitchingDomainForUS(ctx, olt, oc, cmd.Tpid, cmd.Vid, onuID)

				logger.Debug(ctx, fmt.Sprintf("Sequence results: %v, %v, %v, %v", err01, err02, err03, err04))
				if onuDev, ok := onu.(*L2oamOnuDevice); ok {
					onuDev.Base.FlowAdded = true
					onu.updateMap()
				} else {
					logger.Error(ctx, fmt.Sprintf("assertion failed. device-id:%s", onuDeviceID))
				}
			}
		}
	}
	return nil
}

// L2oamAddFlowToDeviceDS runs add-flow-device sequence for downstream
func L2oamAddFlowToDeviceDS(ctx context.Context, dh *DeviceHandler, cmd *L2oamCmd) *l2oam.TomiObjectContext {
	logger.Debug(ctx, "L2oamAddFlowToDeviceDS() Start.")

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamAddFlowToDeviceDS() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
	}

	oc, err01 := sendSetGenericActionCreateForDS(ctx, olt)
	if oc != nil {
		onuID := &l2oam.TomiObjectContext{
			Branch:   0x0c,
			Type:     0x0011,
			Length:   4,
			Instance: 0x00000001,
		}
		err02 := sendSetL2SwitchingDomainForDS(ctx, olt, oc, cmd.Tpid, cmd.Vid)
		err03 := sendSetDefaultOutlet(ctx, olt, oc, onuID)
		logger.Debug(ctx, fmt.Sprintf("Sequence results: %v, %v, %v", err01, err02, err03))
	} else {
		logger.Debug(ctx, fmt.Sprintf("Sequence results: %v", err01))
	}
	olt.setL2oamCmd(cmd)
	olt.setObjectContext(oc)
	olt.updateMap()

	return oc
}

// L2oamAddFlowToDeviceUS runs add-flow-device sequence for upstream
func L2oamAddFlowToDeviceUS(ctx context.Context, dh *DeviceHandler, oc *l2oam.TomiObjectContext, cmd *L2oamCmd) {
	logger.Debug(ctx, "L2oamAddFlowToDeviceUS() Start.")

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamAddFlowToDeviceUS() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
	}

	onuID := &l2oam.TomiObjectContext{
		Branch:   0x0c,
		Type:     0x0011,
		Length:   4,
		Instance: 0x00000002,
	}
	err01 := sendSetL2SwitchingDomainForUS(ctx, olt, oc, cmd.Tpid, cmd.Vid, onuID)

	logger.Debug(ctx, fmt.Sprintf("Sequence results: %v", err01))
}

// L2oamRebootDevice reboots the device
func L2oamRebootDevice(ctx context.Context, dh *DeviceHandler, device *voltha.Device) {
	logger.Debug(ctx, "L2oamRebootDevice() Start.")

	L2oamDisableOlt(ctx, dh)

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamRebootDevice() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
		return
	}

	err01 := sendSetActionReset(ctx, olt)

	logger.Debug(ctx, fmt.Sprintf("Sequence results: %v", err01))

}

// L2oamDeleteOlt deletes the OLT device
func L2oamDeleteOlt(ctx context.Context, dh *DeviceHandler) {
	logger.Debug(ctx, fmt.Sprintf("L2oamDeleteOlt() Start. deviceId=%s", dh.device.Id))

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamDeleteOlt() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
		return
	}
	olt.setAutonomousFlag(false)
	if !SetupL2oamDeleteFlag(dh.device.Id) {
		logger.Warn(ctx, fmt.Sprintf("L2oamDeleteOlt() olt deleted deviceId=%s", dh.device.Id))
		return
	}
	err01 := sendSetAdminState(ctx, olt, true)
	DeleteAllDevice()

	logger.Debug(ctx, fmt.Sprintf("L2oamDeleteOlt() End. deviceId=%s, %v", dh.device.Id, err01))
}

// L2oamDisableOlt disables the OLT device
func L2oamDisableOlt(ctx context.Context, dh *DeviceHandler) {
	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamDisableOlt() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
		return
	}
	olt.setAutonomousFlag(false)

	onuDevices, err := dh.coreProxy.GetChildDevices(ctx, dh.device.Id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("catnnot-get-child-devices: %v", err))
	}

	if err = olt.removeFlow(ctx); err != nil {
		logger.Error(ctx, fmt.Sprintf("failed-to-remove-flows: %v", err))
	}
	L2oamRemoveFlow(ctx, dh, onuDevices)
	L2oamChildDeviceLost(ctx, dh, onuDevices)
	L2oamDisableDevice(ctx, dh)
	resetFlowMap()
}

// L2oamRemoveFlowFromDevice removes flows from the device
func L2oamRemoveFlowFromDevice(ctx context.Context, dh *DeviceHandler, onuDevices *voltha.Devices) {
	logger.Debug(ctx, "L2oamRemoveFlowFromDeviceDS() Start.")

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)

	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamRemoveFlowFromDeviceDS() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
	}
	oc := olt.getObjectContext()
	cmd := olt.getL2oamCmd()
	if cmd == nil {
		logger.Warn(ctx, fmt.Sprintf("L2oamRemoveFlowFromDeviceDS() tpid, vid are not specified. deviceId=%s",
			dh.device.Id))
		return
	}
	logger.Debug(ctx, fmt.Sprintf("TPID:%v, VID:%v", cmd.Tpid, cmd.Vid))

	for _, onuDevice := range onuDevices.Items {
		//onu := FindL2oamDevice(onuDevice.MacAddress)
		onu := FindL2oamDeviceByDeviceID(onuDevice.Id)
		if onu == nil {
			logger.Error(ctx, fmt.Sprintf("catnnot-get-onu-device. device-id:%s, mac-address:%s",
				onuDevice.Id, onuDevice.MacAddress))
			continue
		}
		var onuDev *L2oamOnuDevice
		var ok bool
		if onuDev, ok = onu.(*L2oamOnuDevice); ok {
			if !onuDev.Base.FlowAdded {
				logger.Error(ctx, fmt.Sprintf("flow is not added. device-id:%s, mac-address:%s",
					onuDevice.Id, onuDevice.MacAddress))
				continue
			}
		} else {
			logger.Error(ctx, fmt.Sprintf("assertion failed. device-id:%s, mac-address:%s",
				onuDevice.Id, onuDevice.MacAddress))
			continue
		}
		onuID := onu.getObjectContext()
		if onuID == nil {
			logger.Error(ctx, fmt.Sprintf("catnnot-get-onu-object-context. device-id:%s, mac-address:%s",
				onuDevice.Id, onuDevice.MacAddress))
			continue
		}
		logger.Debug(ctx, fmt.Sprintf("Switching Domain/Action Remove Inlet entry. device-id:%s, instance-id:%d",
			onuDevice.Id, onuID.Instance))
		err := sendSetActionInletEntryUsDel(ctx, olt, oc, cmd.Tpid, cmd.Vid, onuID)
		logger.Debug(ctx, fmt.Sprintf("Sequence results: %v", err))
		onuDev.Base.FlowAdded = false
		onu.updateMap()
	}

	if oltDev, ok := olt.(*L2oamOltDevice); ok {
		if oltDev.Base.FlowAdded {
			err02 := sendSetActionInletEntryDsDel(ctx, olt, oc, cmd.Tpid, cmd.Vid)
			err03 := sendSetActionDeleteStream(ctx, olt, oc)

			logger.Debug(ctx, fmt.Sprintf("Sequence results: %v, %v", err02, err03))
			olt.setObjectContext(nil)
			oltDev.Base.FlowAdded = false
			olt.updateMap()
		} else {
			logger.Debug(ctx, fmt.Sprintf("flow is not added. device-id:%s", dh.device.Id))
		}
	} else {
		logger.Debug(ctx, fmt.Sprintf("assersion failed. device-id:%s", dh.device.Id))
	}
}

// L2oamRemoveFlow removes flows
func L2oamRemoveFlow(ctx context.Context, dh *DeviceHandler, onuDevices *voltha.Devices) {
	logger.Debug(ctx, "L2oamRemoveFlow() Start.")

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamRemoveFlow() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
	}

	if flowVersion == 2 {
		//refTbl := olt.getReferenceTable()
		downProfile, upProfile := olt.getProfile()
		//if refTbl == nil || downProfile == nil || upProfile == nil {
		if downProfile == nil || upProfile == nil {
			logger.Warn(ctx, fmt.Sprintf("L2oamRemoveFlow() profiles are not specified. deviceId=%s",
				dh.device.Id))
			return
		}
		downInstanceID := downProfile.GetTrafficProfile()
		upInstanceID := upProfile.GetTrafficProfile()
		//err01 := sendSetTrafficProfile(ctx, olt,
		//	refTbl.GetReferenceControlDown(), downInstanceId)
		//err02 := sendSetTrafficProfile(ctx, olt,
		//	refTbl.GetReferenceControlUp(), upInstanceId)
		err03 := sendSetActionDelete(ctx, olt, downInstanceID, l2oam.ActionTypeTrafficProfile)
		err04 := sendSetActionDelete(ctx, olt, upInstanceID, l2oam.ActionTypeTrafficProfile)
		olt.setProfile(nil, nil)
		olt.updateMap()

		//logger.Debug(ctx, fmt.Sprintf("Sequence results: %v, %v, %v, %v", err01, err02, err03, err04))
		logger.Debug(ctx, fmt.Sprintf("Sequence results: %v, %v", err03, err04))
	} else if flowVersion == 3 {
		for _, onuDevice := range onuDevices.Items {
			logger.Debugf(ctx, "remove-traffic-schedule. onu-device-id:%s", onuDevice.Id)
			onu := FindL2oamDeviceByDeviceID(onuDevice.Id)
			if onu == nil {
				logger.Errorf(ctx, "catnnot-get-onu-device")
				continue
			}
			refTbl := onu.getReferenceTable()
			if refTbl == nil {
				logger.Debug(ctx, "traffic-profile-is-not-created")
				continue
			}
			downProfile, upProfile := onu.getProfile()
			if downProfile == nil || upProfile == nil {
				logger.Debug(ctx, "flow is not added")
				continue
			}
			downInstanceID := downProfile.GetTrafficProfile()
			upInstanceID := upProfile.GetTrafficProfile()
			err01 := sendSetTrafficProfile(ctx, olt,
				refTbl.GetReferenceControlDown(), downInstanceID)
			err02 := sendSetTrafficProfile(ctx, olt,
				refTbl.GetReferenceControlUp(), upInstanceID)
			err03 := sendSetActionDelete(ctx, olt, downInstanceID, l2oam.ActionTypeTrafficProfile)
			err04 := sendSetActionDelete(ctx, olt, upInstanceID, l2oam.ActionTypeTrafficProfile)
			onu.setReferenceTable(nil)
			onu.setProfile(nil, nil)
			onu.updateMap()
			logger.Debug(ctx, fmt.Sprintf("remove-traffic-schedule-results: %v, %v, %v, %v", err01, err02, err03, err04))
		}
	}
}

// L2oamChildDeviceLost sends DeviceLost message
func L2oamChildDeviceLost(ctx context.Context, dh *DeviceHandler, onuDevices *voltha.Devices) {
	logger.Debug(ctx, "L2oamChildDeviceLost() Start.")
	if onuDevices == nil {
		return
	}

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamChildDeviceLost() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
	}

	for _, onuDevice := range onuDevices.Items {
		//onu := FindL2oamDevice(onuDevice.MacAddress)
		onu := FindL2oamDeviceByDeviceID(onuDevice.Id)
		if onu == nil {
			logger.Error(ctx, fmt.Sprintf("catnnot-get-onu-device. device-id:%s, mac-address:%s",
				onuDevice.Id, onuDevice.MacAddress))
		} else {
			oc := onu.getObjectContext()
			if oc == nil {
				logger.Error(ctx, fmt.Sprintf("catnnot-get-onu-object-context. device-id:%s, mac-address:%s",
					onuDevice.Id, onuDevice.MacAddress))
			} else {
				logger.Debug(ctx, fmt.Sprintf("onu-object-context. instance-id:%x", oc.Instance))
				err01 := sendSetActionDeleteOnu(ctx, olt, oc)

				logger.Debug(ctx, fmt.Sprintf("Sequence results: %v", err01))
			}
		}
	}
}

// L2oamDisableDevice disables the device
func L2oamDisableDevice(ctx context.Context, dh *DeviceHandler) {
	logger.Debug(ctx, "L2oamDisableDevice() Start.")

	olt := FindL2oamDeviceByDeviceID(dh.device.Id)
	if olt == nil {
		logger.Error(ctx, fmt.Sprintf("L2oamDisableDevice() FindL2oamDeviceByDeviceId() error. olt not found. deviceId=%s", dh.device.Id))
	}

	oltDev := olt.(*L2oamOltDevice)
	if oltDev.Base.ActionIds != nil {
		var bytes [4]byte
		if id, ok := oltDev.Base.ActionIds[ActionIDFilterPonPort]; ok {
			binary.BigEndian.PutUint32(bytes[:], id)
			if err := sendSetActionDelete(ctx, olt, bytes[:], l2oam.ActionTypeProtocolFilter); err != nil {
				return
			}
		}
		if id, ok := oltDev.Base.ActionIds[ActionIDFilterEthPort]; ok {
			binary.BigEndian.PutUint32(bytes[:], id)
			if err := sendSetActionDelete(ctx, olt, bytes[:], l2oam.ActionTypeProtocolFilter); err != nil {
				return
			}
		}
		oltDev.Base.ActionIds = nil
		olt.updateMap()
	}
	err01 := sendSetAdminState(ctx, olt, true)
	err02 := sendSetManagementLock(ctx, olt)

	logger.Debug(ctx, fmt.Sprintf("Sequence results: %v, %v", err01, err02))
}

func sendDiscoverySOLICIT(ctx context.Context, device L2oamDevice) error {
	if err := device.send(l2oam.GenerateDiscoverySOLICIT()); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendDiscoverySOLICIT() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.DiscoveryHello{}
	if err := packet.Decode(ether.Payload); err != nil {
		return nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendDiscoverySOLICIT() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return nil
}

func sendGetVendorName(ctx context.Context, device L2oamDevice) (string, error) {
	if err := device.send(l2oam.GnerateGetVendorName()); err != nil {
		return "", nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return "", fmt.Errorf("[%s] sendGetVendorName() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.GetVendorNameRes{}
	if err := packet.Decode(ether.Payload); err != nil {
		return "", nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendGetVendorName() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return packet.GetVendorName(), nil
}

func sendGetModuleNumber(ctx context.Context, device L2oamDevice) (string, error) {
	if err := device.send(l2oam.GenerateGetModuleNumber()); err != nil {
		return "", nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return "", fmt.Errorf("[%s] sendGetModuleNumber() Send Error: %v", device.getDeviceName(), err)
	}
	packet := &l2oam.GetModuleNumberRes{}
	if err := packet.Decode(ether.Payload); err != nil {
		return "", nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendGetModuleNumber() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return packet.GetModuleNumber(), nil
}

func sendManuFacturerInfo(ctx context.Context, device L2oamDevice) (string, error) {
	if err := device.send(l2oam.GenerateManuFacturerInfo()); err != nil {
		return "", nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return "", fmt.Errorf("[%s] sendManuFacturerInfo() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.GetManufacturerRes{}
	if err := packet.Decode(ether.Payload); err != nil {
		return "", nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendManuFacturerInfo() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return packet.GetManufacturer(), nil
}

func sendGetFirmwareVersion(ctx context.Context, device L2oamDevice) (string, error) {
	if err := device.send(l2oam.GenerateGetFirmwareVersion()); err != nil {
		return "", nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return "", fmt.Errorf("[%s] sendGetFirmwareVersion() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.GetFirmwareVersionRes{}
	if err := packet.Decode(ether.Payload); err != nil {
		return "", nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendGetFirmwareVersion() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return packet.GetFirmwareVersionNumber(), nil
}

func sendGetMacAddress(ctx context.Context, device L2oamDevice) (string, error) {
	if err := device.send(l2oam.GenerateGetMacAddress()); err != nil {
		return "", nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return "", fmt.Errorf("[%s] sendGetMacAddress() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.GetMacAddressRes{}
	if err := packet.Decode(ether.Payload); err != nil {
		return "", nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendGetMacAddress() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return packet.GetMacAddress(), nil
}

func sendGetSerialNumber(ctx context.Context, device L2oamDevice) (string, error) {
	if err := device.send(l2oam.GenerateGetSerialNumber()); err != nil {
		return "", nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return "", fmt.Errorf("[%s] sendGetSerialNumber() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.GetSerialNumberRes{}
	if err := packet.Decode(ether.Payload); err != nil {
		return "", nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendGetSerialNumber() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return packet.GetSerialNumber(), nil
}

func sendSetHbtxTemplate(ctx context.Context, device L2oamDevice) error {
	if err := device.send(l2oam.GenerateSetHbtxTemplate()); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetHbtxTemplate() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.TOAMSetResponse{}
	if err := packet.Decode(ether.Payload); err != nil {
		return nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetHbtxTemplate() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return nil
}

func sendGetPonMode(ctx context.Context, device L2oamDevice) (string, error) {
	if err := device.send(l2oam.GenerateGetPonMode()); err != nil {
		return "", nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return "", fmt.Errorf("[%s] sendGetPonMode() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.GetPonMode{}
	if err := packet.Decode(ether.Payload); err != nil {
		return "", nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendGetPonMode() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return packet.GetPonMode(), nil
}

func sendSetMPCPSync(ctx context.Context, device L2oamDevice) error {
	if err := device.send(l2oam.GenerateSetMPCPSync()); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetMPCPSync() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.TOAMSetResponse{}
	if err := packet.Decode(ether.Payload); err != nil {
		return nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetMPCPSync() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return nil
}

func sendSetAdminState(ctx context.Context, device L2oamDevice, isDelete bool) error {
	if isDelete {
		if err := device.send(l2oam.GenerateSetAdminStateDelete()); err != nil {
			return nil
		}
	} else {
		if err := device.send(l2oam.GenerateSetAdminState()); err != nil {
			return nil
		}
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetAdminState() Send Error: isDelete=%v, %v", device.getDeviceName(), isDelete, err)
	}

	packet := &l2oam.TOAMSetResponse{}
	if err := packet.Decode(ether.Payload); err != nil {
		return nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetAdminState() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return nil
}

func sendSetHbtxPeriod(ctx context.Context, device L2oamDevice) error {
	if err := device.send(l2oam.GenerateSetHbtxPeriod()); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetHbtxPeriod() Send Error: %v", device.getDeviceName(), err)
	}
	packet := &l2oam.TOAMSetResponse{}
	if err := packet.Decode(ether.Payload); err != nil {
		return nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetHbtxPeriod() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

	return nil
}

func sendSetIngressPort(ctx context.Context, device L2oamDevice, actionID uint32, isPonPort bool) error {
	if err := device.send(l2oam.GenerateIngressPort(actionID, isPonPort)); err != nil {
		return nil
	}
	_, err := device.waitResponse(ResponseTimer)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("sendSetIngressPort() send error. %s ", err))
		return nil
	}

	return nil
}

func sendSetCaptureProtocols(ctx context.Context, device L2oamDevice, actionID uint32) error {
	if err := device.send(l2oam.GenerateCaptureProtocols(actionID)); err != nil {
		return nil
	}
	_, err := device.waitResponse(ResponseTimer)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("sendSetCaptureProtocols() send error. %s ", err))
		return nil
	}

	return nil
}

func sendGetTrafficControlReferenceTable(ctx context.Context, device L2oamDevice, oc *l2oam.TomiObjectContext) *l2oam.GetTrafficControlReferenceTableRes {
	if err := device.send(l2oam.GenerateGetTrafficControlReferenceTableReq(oc)); err != nil {
		return nil
	}
	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("sendGetTrafficControlReferenceTable() send error. %s ", err))
		return nil
	}
	packet := &l2oam.GetTrafficControlReferenceTableRes{}
	if err := packet.Decode(ether.Payload); err != nil {
		return nil
	}
	return packet
}

func sendSetGenericActionCreate(ctx context.Context, device L2oamDevice, actionType int) *l2oam.SetGenericActionCreateRes {
	if err := device.send(l2oam.GenerateGenericActionCreate(actionType)); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("[%s] sendSetGenericActionCreate() Send Error: %v", device.getDeviceName(), err))
		return nil
	}
	packet := &l2oam.SetGenericActionCreateRes{}
	if err := packet.Decode(ether.Payload); err != nil {
		return nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetGenericActionCreate() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

	return packet
}

func sendSetTrafficControl(ctx context.Context, device L2oamDevice, trafficControl []byte, trafficProfile []byte) error {

	if err := device.send(l2oam.GenerateTrafficControlTrafficProfile(trafficControl, trafficProfile)); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetTrafficControl() Send Error: %v", device.getDeviceName(), err)
	}
	packet := &l2oam.TOAMSetResponse{}

	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetTrafficControl() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

	return nil
}

func sendSetPriority(ctx context.Context, device L2oamDevice, trafficProfile []byte) error {
	if err := device.send(l2oam.GeneratePriority(trafficProfile)); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetPriority() Send Error: %v", device.getDeviceName(), err)
	}
	packet := &l2oam.TOAMSetResponse{}

	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetPriority() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

	return nil
}

func sendSetGuranteedRate(ctx context.Context, device L2oamDevice, cir []byte, trafficProfile []byte) error {
	if err := device.send(l2oam.GenerateGuaranteedRate(cir, trafficProfile)); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetGuranteedRate() Send Error: %v", device.getDeviceName(), err)
	}
	packet := &l2oam.TOAMSetResponse{}

	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetGuranteedRate() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

	return nil
}

// func sendSetGuranteedMbs(ctx context.Context, device L2oamDevice) error {
// 	if err := device.send(l2oam.GenerateGuaranteedMbs()); err != nil {
// 		return nil
// 	}

// 	ether, err := device.waitResponse(ResponseTimer)
// 	if err != nil {
// 		return fmt.Errorf("[%s] sendSetGuranteedMbs() Send Error: %v", device.getDeviceName(), err)
// 	}
// 	packet := &l2oam.TOAMSetResponse{}
// 	//packet.Decode(ether.Payload)

// 	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetGuranteedMbs() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

// 	return nil
// }

func sendSetBestEffortRate(ctx context.Context, device L2oamDevice, pir []byte, trafficProfile []byte) error {
	if err := device.send(l2oam.GenerateBestEffortRate(pir, trafficProfile)); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetBestEffortRate() Send Error: %v", device.getDeviceName(), err)
	}
	packet := &l2oam.TOAMSetResponse{}

	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetBestEffortRate() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

	return nil
}

// func sendSetBestEffortMbs(ctx context.Context, device L2oamDevice) error {
// 	if err := device.send(l2oam.GenerateBestEffortMbs()); err != nil {
// 		return nil
// 	}

// 	ether, err := device.waitResponse(ResponseTimer)
// 	if err != nil {
// 		return fmt.Errorf("[%s] sendSetBestEffortMbs() Send Error: %v", device.getDeviceName(), err)
// 	}
// 	packet := &l2oam.TOAMSetResponse{}
// 	//packet.Decode(ether.Payload)

// 	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetBestEffortMbs() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

// 	return nil
// }

func sendSetGenericActionCreateForDS(ctx context.Context, device L2oamDevice) (*l2oam.TomiObjectContext, error) {
	if err := device.send(l2oam.GenerateGenericActionCreateDs()); err != nil {
		return nil, nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return nil, fmt.Errorf("[%s] sendSetGenericActionCreateForDS() Send Error: %v", device.getDeviceName(), err)
	}
	oc := &l2oam.TomiObjectContext{}

	i := 29
	oc.Branch = ether.Payload[i]
	i++
	oc.Type = binary.BigEndian.Uint16(ether.Payload[i : i+2])
	i += 2
	oc.Length = ether.Payload[i]
	i++
	oc.Instance = binary.BigEndian.Uint32(ether.Payload[i : i+4])

	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetGenericActionCreateForDS() Success, response=%x, branch=%x, type=%x, length=%x, instance=%x",
		device.getDeviceName(), ether.Payload, oc.Branch, oc.Type, oc.Length, oc.Instance))

	return oc, nil
}

func sendSetL2SwitchingDomainForDS(ctx context.Context, device L2oamDevice, oc *l2oam.TomiObjectContext, tpid []byte, vid []byte) error {
	tibitData := l2oam.GenerateL2switchingDomainActionLnletEntryDs(oc, tpid, vid)
	if err := device.send(tibitData); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetL2SwitchingDomain() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.TOAMSetResponse{}

	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetL2SwitchingDomain() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

	return nil
}

func sendSetDefaultOutlet(ctx context.Context, device L2oamDevice, oc *l2oam.TomiObjectContext, onuID *l2oam.TomiObjectContext) error {
	tibitData := l2oam.GenerateSetDefaultOutlet(oc, onuID)
	if err := device.send(tibitData); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetDefaultOutlet() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.TOAMSetResponse{}

	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetDefaultOutlet() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

	return nil
}

func sendSetL2SwitchingDomainForUS(ctx context.Context, device L2oamDevice, oc *l2oam.TomiObjectContext,
	tpid []byte, vid []byte, onuID *l2oam.TomiObjectContext) error {
	tibitData := l2oam.GenerateL2switchingDomainActionLnletEntryUs(oc, tpid, vid, onuID)
	if err := device.send(tibitData); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetL2SwitchingDomainForUS() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.TOAMSetResponse{}

	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetL2SwitchingDomainForUS() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))

	return nil
}

func sendSetActionReset(ctx context.Context, device L2oamDevice) error {
	if err := device.send(l2oam.GenerateSetActionReset()); err != nil {
		return nil
	}

	ether, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetActionReset() Send Error: %v", device.getDeviceName(), err)
	}

	packet := &l2oam.TOAMSetResponse{}
	if err := packet.Decode(ether.Payload); err != nil {
		return nil
	}
	logger.Debug(ctx, fmt.Sprintf("[%s] sendSetActionReset() Success, response=%x:%s", device.getDeviceName(), ether.Payload, packet))
	return nil
}

func sendSetTrafficProfile(ctx context.Context, device L2oamDevice,
	trafficControl []byte, trafficProfile []byte) error {
	if err := device.send(l2oam.GenerateSetTrafficProfile(trafficControl, trafficProfile)); err != nil {
		return err
	}

	if _, err := device.waitResponse(ResponseTimer); err != nil {
		return fmt.Errorf("[%s] sendSetTrafficProfile() Send Error: %v", device.getDeviceName(), err)
	}

	return nil
}

func sendSetActionDelete(ctx context.Context, device L2oamDevice, trafficProfile []byte, actionType int) error {
	if err := device.send(l2oam.GenerateSetActionDelete(trafficProfile, actionType)); err != nil {
		return nil
	}

	_, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetActionDelete() Send Error: %v", device.getDeviceName(), err)
	}

	return nil
}

func sendSetActionDeleteOnu(ctx context.Context, device L2oamDevice, oc *l2oam.TomiObjectContext) error {
	logger.Debug(ctx, "sendSetActionDeleteOnu() Start.")
	if err := device.send(l2oam.GenerateSetActionDeleteOnu(oc)); err != nil {
		return nil
	}

	_, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetActionDeleteOnu() Send Error: %v", device.getDeviceName(), err)
	}

	return nil
}

func sendSetManagementLock(ctx context.Context, device L2oamDevice) error {
	if err := device.send(l2oam.GenerateSetManagementLock()); err != nil {
		return nil
	}

	_, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetManagementLock() Send Error: %v", device.getDeviceName(), err)
	}

	return nil
}

func sendSetActionInletEntryDsDel(ctx context.Context, device L2oamDevice,
	oc *l2oam.TomiObjectContext, tpid []byte, vid []byte) error {
	if err := device.send(l2oam.GenerateL2switchingDomainActionInletEntryDsDel(oc, tpid, vid)); err != nil {
		return nil
	}

	_, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetActionInletEntryDsDel() Send Error: %v", device.getDeviceName(), err)
	}

	return nil
}

func sendSetActionInletEntryUsDel(ctx context.Context, device L2oamDevice,
	oc *l2oam.TomiObjectContext, tpid []byte, vid []byte, onuID *l2oam.TomiObjectContext) error {
	if err := device.send(l2oam.GenerateL2switchingDomainActionInletEntryUsDel(oc, tpid, vid, onuID)); err != nil {
		return nil
	}

	_, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetActionInletEntryUsDel() Send Error: %v", device.getDeviceName(), err)
	}

	return nil
}

func sendSetActionDeleteStream(ctx context.Context, device L2oamDevice, oc *l2oam.TomiObjectContext) error {
	if err := device.send(l2oam.GenerateSetActionDeleteStream(oc)); err != nil {
		return nil
	}

	_, err := device.waitResponse(ResponseTimer)
	if err != nil {
		return fmt.Errorf("[%s] sendSetActionDeleteStream() Send Error: %v", device.getDeviceName(), err)
	}

	return nil
}
