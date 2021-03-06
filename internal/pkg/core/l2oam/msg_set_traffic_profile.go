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

// GenerateSetTrafficProfile generates "Traffic Control/Traffic Profile" message
func GenerateSetTrafficProfile(trafficControl []byte, trafficProfile []byte) gopacket.SerializableLayer {
	tibitData := &setTrafficControlTrafficProfileReq{
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
		OCBranch:   0x0c,
		OCType:     0x07c0,
		OCLength:   4,
		OCInstance: trafficControl,
		// Vc
		VcBranch: 0x7c,
		VcLeaf:   0x0002,
		VcLength: 9,
		// EC OC
		ECOC: []ECOCTrafficControlTrafficProfile{
			{
				EcOcLength:   8,
				EcOcBranch:   0x0c,
				EcOcType:     0x070f,
				EcOcLength2:  4,
				EcOcInstance: trafficProfile,
			},
		},
		// End
		EndBranch: 0x00,
	}
	return tibitData
}
