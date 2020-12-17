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

// GenerateGenericActionCreateDs generates "Generic/Action Create" message
func GenerateGenericActionCreateDs() gopacket.SerializableLayer {

	tibitData := &setGenericActionCreateReq{
		// IEEE 1904.2
		Opcode: 0x03,
		// OMI Protocol
		Flags:      0x0050,
		OAMPDUCode: 0xfe,
		OUId:       []byte{0x2a, 0xea, 0x15},
		// TiBiT OLT Management Interface
		TOMIOpcode: 0x03,
		// Correlation Tag
		CTBranch:   0x0c,
		CTType:     0x0c7a,
		CTLength:   4,
		CTInstance: getOltInstance(),
		// Object Context
		OCBranch:   0x0c,
		OCType:     0x0dce,
		OCLength:   4,
		OCInstance: 0x00000000,
		//Vc
		VcBranch: 0x6e,
		VcLeaf:   0x7001,
		VcLength: 4,
		//OT
		OT: []OTGenericActionCreate{{OTLength: 3, OTValue: []byte{0x0c, 0x02, 0x5d}}},
		// End
		EndBranch: 0,
	}

	return tibitData

}
