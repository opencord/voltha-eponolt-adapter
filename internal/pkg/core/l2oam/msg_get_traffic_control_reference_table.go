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
	"github.com/google/gopacket/layers"
)

// GenerateGetTrafficControlReferenceTableReq generates "PON Link/Traffic Control Reference Table" message
func GenerateGetTrafficControlReferenceTableReq(oc *TomiObjectContext) gopacket.SerializableLayer {
	data := &TOAMGetRequest{
		// IEEE 1904.2
		Opcode: 0x03,
		// OAM Protocol
		Flags:      0x0050,
		OAMPDUCode: 0xfe,
		OUId:       []byte{0x2a, 0xea, 0x15},
		// TiBit OLT Management Interface
		TOMIOpcode: 0x01,
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
		// Vd
		VdBranch: 0x01,
		VdLeaf:   0x0007,
		// End
		EndBranch: 0x00,
	}

	return data
}

// GetTrafficControlReferenceTableRes is a structure for a response of "Traffic Control Reference Table" message
type GetTrafficControlReferenceTableRes struct {
	layers.BaseLayer
	Opcode uint8
	Flags      uint16
	OAMPDUCode uint8
	OUId       []byte // Organizationally Unique Identifier: 2a:ea:15 (Tibit Communications)
	TOMIOpcode uint8
	CTBranch   uint8
	CTType     uint16
	CTLength   uint8
	CTInstance uint32
	OCBranch   uint8
	OCType     uint16
	OCLength   uint8
	OCInstance uint32
	VcBranch uint8
	VcLeaf   uint16
	VcLength uint8

	EcOcLengthDown   uint8
	EcOcBranchDown   uint8
	EcOcTypeDown     uint16
	EcOcLength2Down  uint8
	EcOcInstanceDown []byte

	EcOcLengthUp   uint8
	EcOcBranchUp   uint8
	EcOcTypeUp     uint16
	EcOcLength2Up  uint8
	EcOcInstanceUp []byte

	EndBranch uint8
}

// Decode decodes byte arrays to a data structure
func (d *GetTrafficControlReferenceTableRes) Decode(data []byte) error {
	i := 0
	d.Opcode = data[i]
	i++
	d.Flags = binary.BigEndian.Uint16(data[i : i+2])
	i += 2
	d.OAMPDUCode = data[i]
	i++
	d.OUId = data[i : i+3]
	i += len(d.OUId)
	d.TOMIOpcode = data[i]
	i++
	d.CTBranch = data[i]
	i++
	d.CTType = binary.BigEndian.Uint16(data[i : i+2])
	i += 2
	d.CTLength = data[i]
	i++
	d.CTInstance = binary.BigEndian.Uint32(data[i : i+4])
	i += 4
	d.OCBranch = data[i]
	i++
	d.OCType = binary.BigEndian.Uint16(data[i : i+2])
	i += 2
	d.OCLength = data[i]
	i++
	d.OCInstance = binary.BigEndian.Uint32(data[i : i+4])
	i += 4
	d.VcBranch = data[i]
	i++
	d.VcLeaf = binary.BigEndian.Uint16(data[i : i+2])
	i += 2
	d.VcLength = data[i]
	i++

	d.EcOcLengthDown = data[i]
	i++
	d.EcOcBranchDown = data[i]
	i++
	d.EcOcTypeDown = binary.BigEndian.Uint16(data[i : i+2])
	i += 2
	d.EcOcLength2Down = data[i]
	i++
	d.EcOcInstanceDown = data[i : i+int(d.EcOcLength2Down)]
	i += int(d.EcOcLength2Down)

	d.EcOcLengthUp = data[i]
	i++
	d.EcOcBranchUp = data[i]
	i++
	d.EcOcTypeUp = binary.BigEndian.Uint16(data[i : i+2])
	i += 2
	d.EcOcLength2Up = data[i]
	i++
	d.EcOcInstanceUp = data[i : i+int(d.EcOcLength2Up)]
	i += int(d.EcOcLength2Up)

	d.EndBranch = data[i]

	return nil
}

// GetReferenceControlDown returns a link id for downstream
func (d *GetTrafficControlReferenceTableRes) GetReferenceControlDown() []byte {
	return d.EcOcInstanceDown
}

// GetReferenceControlUp returns a link id for upstream
func (d *GetTrafficControlReferenceTableRes) GetReferenceControlUp() []byte {
	return d.EcOcInstanceUp
}
