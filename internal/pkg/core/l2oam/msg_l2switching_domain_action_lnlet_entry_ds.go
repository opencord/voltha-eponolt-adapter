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

// GenerateL2switchingDomainActionLnletEntryDs generates "L2 Switching Domain/Action Add Inlet entry(for downstream)"
func GenerateL2switchingDomainActionLnletEntryDs(oc *TomiObjectContext, tmTpidValue []byte, tmVidValue []byte) gopacket.SerializableLayer {

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
		//VcLength: uint8(24 + len(tmTpidValue) + len(tmVidValue)), //27
		VcLength: uint8(23 + len(tmTpidValue) + len(tmVidValue)), //26
		// Source OC
		SOLength:    5,
		SOBranch:    0x0c,
		SOType:      0x0eca,
		SOValLength: 1,
		SOInstance:  0x00,
		// TagMatchList
		//TMLLength: uint8(7 + len(tmTpidValue) + len(tmVidValue)), //10
		TMLLength: uint8(6 + len(tmTpidValue) + len(tmVidValue)), //9
		TMLList: []TpidVid{
			{Length: uint8(2 + len(tmTpidValue) + len(tmVidValue)),
				TpIDLength: uint8(len(tmTpidValue)),
				TpIDValue:  tmTpidValue,
				VIdLength:  uint8(len(tmVidValue)),
				VIdValue:   tmVidValue},
			////{Length: 5, TpIdLength: 2, TpIdValue: []byte{0x88, 0xa8}, VIdLength: 1, VIdValue: []byte{0x64}}

			//{Length: 3, TpIdLength: 128, TpIdValue: []byte{0}, VIdLength: 1, VIdValue: []byte{0x00}}},
			{Length: 2, TpIDLength: 128, TpIDValue: []byte{0}, VIdLength: 128, VIdValue: []byte{0x00}}},
		// TagOpPop(UI)
		TOPopLength: 1,
		TOPopValue:  0x01,
		// TagOpSetList
		TOSLength: 3,
		TOSList: []TpidVid{
			{Length: 2, TpIDLength: 128, TpIDValue: []byte{0}, VIdLength: 128, VIdValue: []byte{0}}},
		// TagOpPushList
		TOPushLength: 3,
		TOPushList: []TpidVid{
			{Length: 2, TpIDLength: 128, TpIDValue: []byte{0}, VIdLength: 128, VIdValue: []byte{0}}},
		// End
		EndBranch: 0x00,
	}

	return tibitData
}
