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

// GenerateGuaranteedMbs generates "Traffic Profile/Guaranteed MBS" message
func GenerateGuaranteedMbs() gopacket.SerializableLayer {

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
		OCBranch:   0x0c,
		OCType:     0x070f,
		OCLength:   4,
		OCInstance: 0x00000002,
		// Vc
		VcBranch: 0x7f,
		VcLeaf:   0x0007,
		VcLength: 5,
		// EC
		ECLength: 4,
		ECValue:  []byte{0x00, 0x06, 0x40, 0x00},
		// End
		EndBranch: 0x00,
	}
	return tibitData
}
