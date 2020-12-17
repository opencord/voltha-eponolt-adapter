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

package l2oam

import (
	"github.com/google/gopacket"
)

// GenerateL2switchingDomainActionLnletEntryUs generates "L2 Switching Domain/Action Add Inlet entry(upstream)" message
func GenerateL2switchingDomainActionLnletEntryUs(oc *TomiObjectContext,
	pushTpidValue []byte, pushVidValue []byte, onuID *TomiObjectContext) gopacket.SerializableLayer {

	tibitData := &L2switchingDomainRequest{
		// IEEE 1904.2
		Opcode: 0x03,
		// OAM Protocol
		Flags:      0x0050,
		OAMPDUCode: 0xfe,
		OUId:       []byte{0x2a, 0xea, 0x15},
		// TiBit OLT Management Interface
		TOMIOpcode: 0x03,
		// Correlation Tag
		CTBranch:   0x0c,
		CTType:     0x0c7a,
		CTLength:   4,
		CTInstance: getOltInstance(),
		// Object Context
		OCBranch:   oc.Branch,
		OCType:     oc.Type,
		OCLength:   oc.Length,
		OCInstance: oc.Instance,
		// Vc
		VcBranch: 0x5d,
		VcLeaf:   0x7001,
		VcLength: uint8(24 + len(pushTpidValue) + len(pushVidValue)), //27
		// Source OC
		SOLength:    5,
		SOBranch:    0x0c,
		SOType:      0x0011,
		SOValLength: 1,
		SOInstance:  uint8(onuID.Instance),
		// TagMatchList
		TMLLength: 7,
		TMLList: []TpidVid{
			{
				Length:     3,
				TpIDLength: 0x80,
				TpIDValue:  []byte{0},
				VIdLength:  1,
				VIdValue:   []byte{0x00},
			},
			{
				Length:     2,
				TpIDLength: 0x80,
				TpIDValue:  []byte{0},
				VIdLength:  0x80,
				VIdValue:   []byte{0},
			}},
		// TagOpPop(UI)
		TOPopLength: 1,
		TOPopValue:  0x00,
		// TagOpSetList
		TOSLength: 3,
		TOSList: []TpidVid{
			{
				Length:     2,
				TpIDLength: 0x80,
				TpIDValue:  []byte{0},
				VIdLength:  0x80,
				VIdValue:   []byte{0},
			}},
		// TagOpPushList
		TOPushLength: uint8(3 + len(pushTpidValue) + len(pushVidValue)), //6
		TOPushList: []TpidVid{
			{Length: uint8(2 + len(pushTpidValue) + len(pushVidValue)),
				TpIDLength: uint8(len(pushTpidValue)),
				TpIDValue:  pushTpidValue,
				VIdLength:  uint8(len(pushVidValue)),
				VIdValue:   pushVidValue}}, ///{Length: 5, TpIdLength: 2, TpIdValue: []byte{0x88, 0xa8}, VIdLength: 1, VIdValue: []byte{0x64}}
		// End
		EndBranch: 0x00,
	}

	return tibitData
}
