/*
 * Copyright 2018-present Open Networking Foundation

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

//Package core provides the utility for olt devices, flows and statistics
package core

import (
	"context"

	"github.com/opencord/voltha-lib-go/v3/pkg/flows"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"github.com/opencord/voltha-openolt-adapter/internal/pkg/olterrors"
	ofp "github.com/opencord/voltha-protos/v3/go/openflow_13"
	"github.com/opencord/voltha-protos/v3/go/voltha"
)

/*=====================================================================

@TODO: Looks like this Flow id concept below is not used anywhere
       Propose to remove the below documentation of Flow Id on confirmation
       of the same

Flow id

    Identifies a flow within a single OLT
    Flow Id is unique per OLT
    Multiple GEM ports can map to same flow id

     13    11              4      0
    +--------+--------------+------+
    | pon id |    onu id    | Flow |
    |        |              | idx  |
    +--------+--------------+------+

    14 bits = 16384 flows (per OLT).

    pon id = 4 bits = 16 PON ports
    onu id = 7 bits = 128 ONUss per PON port
    Flow index = 3 bits = 4 bi-directional flows per ONU
                        = 8 uni-directional flows per ONU


Logical (OF) UNI port number

    OpenFlow port number corresponding to PON UNI

     20        12              4      0
    +--+--------+--------------+------+
    |0 | pon id |    onu id    |uni id|
    +--+--------+--------------+------+

    pon id = 8 bits = 256 PON ports
    onu id = 8 bits = 256 ONUs per PON port

Logical (OF) NNI port number

    OpenFlow port number corresponding to PON NNI

     20                             0
    +--+----------------------------+
    |1 |                    intf_id |
    +--+----------------------------+

    No overlap with UNI port number space


PON OLT (OF) port number

    OpenFlow port number corresponding to PON OLT ports

     31     28                                 0
    +--------+------------------------~~~------+
    |  0x2   |          pon intf id            |
    +--------+------------------------~~~------+
*/

const (
	bitsForUniID = 4
	bitsForONUID = 8
	bitsForPONID = 8
	MaxOnusPerPon = (1 << bitsForONUID)
	MaxPonsPerOlt = (1 << bitsForPONID)
	MaxUnisPerOnu = (1 << bitsForUniID)
	nniUniDiffPos = (bitsForUniID + bitsForONUID + bitsForPONID)
	ponIntfMarkerPos = 28
	ponIntfMarkerValue = 0x2
	bitsforNNIID = 20
	minNniIntPortNum = (1 << bitsforNNIID)
	maxNniPortNum = ((1 << (bitsforNNIID + 1)) - 1)
)

//MinUpstreamPortID value
var MinUpstreamPortID = 0xfffd

//MaxUpstreamPortID value
var MaxUpstreamPortID = 0xfffffffd

var controllerPorts = []uint32{0xfffd, 0x7ffffffd, 0xfffffffd}

//MkUniPortNum returns new UNIportNum based on intfID, inuID and uniID
func MkUniPortNum(ctx context.Context, intfID, onuID, uniID uint32) uint32 {
	var limit = int(onuID)
	if limit > MaxOnusPerPon {
		logger.Warn(ctx, "exceeded-the-max-onus-per-pon")
	}
	return (intfID << (bitsForUniID + bitsForONUID)) | (onuID << bitsForUniID) | uniID
}

//OnuIDFromPortNum returns ONUID derived from portNumber
func OnuIDFromPortNum(portNum uint32) uint32 {
	return (portNum >> bitsForUniID) & (MaxOnusPerPon - 1)
}

//IntfIDFromUniPortNum returns IntfID derived from portNum
func IntfIDFromUniPortNum(portNum uint32) uint32 {
	return (portNum >> (bitsForUniID + bitsForONUID)) & (MaxPonsPerOlt - 1)
}

//UniIDFromPortNum return UniID derived from portNum
func UniIDFromPortNum(portNum uint32) uint32 {
	return (portNum) & (MaxUnisPerOnu - 1)
}

//IntfIDToPortNo returns portId derived from intftype, intfId and portType
func IntfIDToPortNo(intfID uint32, intfType voltha.Port_PortType) uint32 {
	if (intfType) == voltha.Port_ETHERNET_NNI {
		return (1 << nniUniDiffPos) | intfID
	}
	if (intfType) == voltha.Port_PON_OLT {
		return (ponIntfMarkerValue << ponIntfMarkerPos) | intfID
	}
	return 0
}

//PortNoToIntfID returns portnumber derived from interfaceID
func PortNoToIntfID(portno uint32, intfType voltha.Port_PortType) uint32 {
	if (intfType) == voltha.Port_ETHERNET_NNI {
		return (1 << nniUniDiffPos) ^ portno
	}
	if (intfType) == voltha.Port_PON_OLT {
		return (ponIntfMarkerValue << ponIntfMarkerPos) ^ portno
	}
	return 0
}

//IntfIDFromNniPortNum returns Intf ID derived from portNum
func IntfIDFromNniPortNum(ctx context.Context, portNum uint32) (uint32, error) {
	if portNum < minNniIntPortNum || portNum > maxNniPortNum {
		logger.Errorw(ctx, "nniportnumber-is-not-in-valid-range", log.Fields{"portnum": portNum})
		return uint32(0), olterrors.ErrInvalidPortRange
	}
	return (portNum & 0xFFFF), nil
}

//IntfIDToPortTypeName returns port type derived from the intfId
func IntfIDToPortTypeName(intfID uint32) voltha.Port_PortType {
	if ((ponIntfMarkerValue << ponIntfMarkerPos) ^ intfID) < MaxPonsPerOlt {
		return voltha.Port_PON_OLT
	}
	if (intfID & (1 << nniUniDiffPos)) == (1 << nniUniDiffPos) {
		return voltha.Port_ETHERNET_NNI
	}
	return voltha.Port_ETHERNET_UNI
}

//ExtractAccessFromFlow returns AccessDevice information
func ExtractAccessFromFlow(inPort, outPort uint32) (uint32, uint32, uint32, uint32) {
	if IsUpstream(outPort) {
		return inPort, IntfIDFromUniPortNum(inPort), OnuIDFromPortNum(inPort), UniIDFromPortNum(inPort)
	}
	return outPort, IntfIDFromUniPortNum(outPort), OnuIDFromPortNum(outPort), UniIDFromPortNum(outPort)
}

//IsUpstream returns true for Upstream and false for downstream
func IsUpstream(outPort uint32) bool {
	for _, port := range controllerPorts {
		if port == outPort {
			return true
		}
	}
	return (outPort & (1 << nniUniDiffPos)) == (1 << nniUniDiffPos)
}

//IsControllerBoundFlow returns true/false
func IsControllerBoundFlow(outPort uint32) bool {
	for _, port := range controllerPorts {
		if port == outPort {
			return true
		}
	}
	return false
}

//OnuIDFromUniPortNum returns onuId from give portNum information.
func OnuIDFromUniPortNum(portNum uint32) uint32 {
	return (portNum >> bitsForUniID) & (MaxOnusPerPon - 1)
}

//FlowExtractInfo fetches uniport from the flow, based on which it gets and returns ponInf, onuID, uniID, inPort and ethType
func FlowExtractInfo(ctx context.Context, flow *ofp.OfpFlowStats, flowDirection string) (uint32, uint32, uint32, uint32, uint32, uint32, error) {
	var uniPortNo uint32
	var ponIntf uint32
	var onuID uint32
	var uniID uint32
	var inPort uint32
	var ethType uint32

	if flowDirection == "upstream" {
		if uniPortNo = flows.GetChildPortFromTunnelId(flow); uniPortNo == 0 {
			for _, field := range flows.GetOfbFields(flow) {
				if field.GetType() == flows.IN_PORT {
					uniPortNo = field.GetPort()
					break
				}
			}
		}
	} else if flowDirection == "downstream" {
		if uniPortNo = flows.GetChildPortFromTunnelId(flow); uniPortNo == 0 {
			for _, field := range flows.GetOfbFields(flow) {
				if field.GetType() == flows.METADATA {
					for _, action := range flows.GetActions(flow) {
						if action.Type == flows.OUTPUT {
							if out := action.GetOutput(); out != nil {
								uniPortNo = out.GetPort()
							}
							break
						}
					}
				} else if field.GetType() == flows.IN_PORT {
					inPort = field.GetPort()
				} else if field.GetType() == flows.ETH_TYPE {
					ethType = field.GetEthType()
				}
			}
		}
	}

	if uniPortNo == 0 {
		return 0, 0, 0, 0, 0, 0, olterrors.NewErrNotFound("pon-interface", log.Fields{"flow-direction": flowDirection}, nil)
	}

	ponIntf = IntfIDFromUniPortNum(uniPortNo)
	onuID = OnuIDFromUniPortNum(uniPortNo)
	uniID = UniIDFromPortNum(uniPortNo)

	logger.Debugw(ctx, "flow-extract-info-result",
		log.Fields{
			"uniportno": uniPortNo,
			"pon-intf":  ponIntf,
			"onu-id":    onuID,
			"uni-id":    uniID,
			"inport":    inPort,
			"ethtype":   ethType})

	return uniPortNo, ponIntf, onuID, uniID, inPort, ethType, nil
}
