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

// GenerateCaptureProtocols generates "Protocol Filter/Capture Protocols" message
func GenerateCaptureProtocols(ocInstance uint32) gopacket.SerializableLayer {

	return &SetGenerteCaptureProtocolsreq{
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
		OCType:     0x0cff,
		OCLength:   4,
		OCInstance: ocInstance,
		// Vc
		VcBranch: 0xcf,
		VcLeaf:   0x0003,
		VcLength: 5,
		// EC
		ECLength: 4,
		//Protocol
		ProtocolLength: 1,
		ProtocolValue:  0x01,
		//Action
		ActionLength: 1,
		ActionValue:  0x01,
		// End
		EndBranch: 0x00,
	}

}

// SetGenerteCaptureProtocolsreq is a structure for a request of "Protocol Filter/Capture Protocols" message
type SetGenerteCaptureProtocolsreq struct {
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
	ECLength uint8
	ProtocolLength uint8
	ProtocolValue  uint8
	ActionLength uint8
	ActionValue  uint8
	EndBranch uint8
}

// Len returns the length of SetGenerteCaptureProtocolsreq
func (d *SetGenerteCaptureProtocolsreq) Len() int {
	return 21 + int(d.CTLength) + int(d.OCLength) + int(d.VcLength)
}

// LayerType returns the ethernet type of SetGenerteCaptureProtocolsreq
func (d *SetGenerteCaptureProtocolsreq) LayerType() gopacket.LayerType {
	return layers.LayerTypeEthernet
}

// SerializeTo serializes a data structure to byte arrays
func (d *SetGenerteCaptureProtocolsreq) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	plen := int(d.Len())
	data, err := b.PrependBytes(plen)
	if err != nil {
		return err
	}

	i := 0
	data[i] = byte(d.Opcode)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.Flags)
	i += 2
	data[i] = byte(d.OAMPDUCode)
	i++
	copy(data[i:i+len(d.OUId)], d.OUId)
	i += len(d.OUId)
	data[i] = byte(d.TOMIOpcode)
	i++
	data[i] = byte(d.CTBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.CTType)
	i += 2
	data[i] = byte(d.CTLength)
	i++
	binary.BigEndian.PutUint32(data[i:i+4], d.CTInstance)
	i += 4
	data[i] = byte(d.OCBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.OCType)
	i += 2
	data[i] = byte(d.OCLength)
	i++
	binary.BigEndian.PutUint32(data[i:i+4], d.OCInstance)
	i += 4
	data[i] = byte(d.VcBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.VcLeaf)
	i += 2
	data[i] = byte(d.VcLength)
	i++
	data[i] = byte(d.ECLength)
	i++
	data[i] = byte(d.ProtocolLength)
	i++
	data[i] = byte(d.ProtocolValue)
	i++
	data[i] = byte(d.ActionLength)
	i++
	data[i] = byte(d.ActionValue)
	i++
	data[i] = byte(d.EndBranch)

	return nil

}

// Decode decodes byte arrays to a data structure
func (d *SetGenerteCaptureProtocolsreq) Decode(data []byte) error {
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
	d.ECLength = data[i]
	i++
	d.ProtocolLength = data[i]
	i++
	d.ProtocolValue = data[i]
	i++
	d.ActionLength = data[i]
	i++
	d.ActionValue = data[i]
	i++
	d.EndBranch = data[i]

	return nil
}
