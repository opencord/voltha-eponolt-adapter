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
	"encoding/binary"

	"github.com/google/gopacket"
)

// GenerateSetActionDelete generates "Generic/Action Delete" message
func GenerateSetActionDelete(trafficProfile []byte, actionType int) gopacket.SerializableLayer {
	ocInstance := binary.BigEndian.Uint32(trafficProfile)
	objectType := getObjectType(actionType)
	tibitData := &TOAMSetRequest{
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
		OCBranch:   objectType[0],
		OCType:     binary.BigEndian.Uint16(objectType[1:3]),
		OCLength:   4,
		OCInstance: ocInstance,
		// Vc
		VcBranch: 0x6e,
		VcLeaf:   0x7002,
		VcLength: 0x80,
		// EC
		ECLength: 0,
		ECValue:  []byte{},
		// End
		EndBranch: 0x00,
	}
	return tibitData
}

// GenerateSetActionDeleteStream generates "Generic/Action Delete(for stream)" message
func GenerateSetActionDeleteStream(oc *TomiObjectContext) gopacket.SerializableLayer {
	tibitData := &TOAMSetRequest{
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
		VcBranch: 0x6e,
		VcLeaf:   0x7002,
		VcLength: 0x80,
		// EC
		ECLength: 0,
		ECValue:  []byte{},
		// End
		EndBranch: 0x00,
	}
	return tibitData
}
