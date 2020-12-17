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

// GenerateIngressPort generates "Protocol Filter/Ingress Port" message
func GenerateIngressPort(OCInstance uint32, isponport bool) gopacket.SerializableLayer {

	var ponPort uint16 = 0x0007

	if !isponport {
		ponPort = 0x0e07
	}

	return &SetGenerteIngressPortreq{
		//IEEE 1904.2
		Opcode:     0x03,
		Flags:      0x0050,
		OAMPDUCode: 0xfe,
		OUId:       []byte{0x2a, 0xea, 0x15},
		//TiBiT OLT Management Interface
		TOMIOpcode: 0x03,
		//Correlation Tag
		CTBranch:   0x0c,
		CTType:     0x0c7a,
		CTLength:   4,
		CTInstance: getOltInstance(),
		//Object Context
		OCBranch:   0x0c,
		OCType:     0x0cff,
		OCLength:   4,
		OCInstance: OCInstance,
		//Vc
		VcBranch: 0xcf,
		VcLeaf:   0x0002,
		VcLength: 9,
		//EC OC
		ECOCLength1:  8,
		ECOCBranch:   0x0c,
		ECOCType:     ponPort,
		ECOCLength2:  4,
		ECOCInstance: 0x00000000,
		//End
		EndBranch: 0x00,
	}

}

// SetGenerteIngressPortreq is a structure for a reqest of "Protocol Filter/Ingress Port" message
type SetGenerteIngressPortreq struct {
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
	ECOCLength1  uint8
	ECOCBranch   uint8
	ECOCType     uint16
	ECOCLength2  uint8
	ECOCInstance uint32
	EndBranch uint8
}

// Len returns the length of SetGenerteIngressPortreq
func (d *SetGenerteIngressPortreq) Len() int {
	return 21 + int(d.CTLength) + int(d.OCLength) + int(d.VcLength)
}

// LayerType returns the ethernet type of SetGenerteIngressPortreq
func (d *SetGenerteIngressPortreq) LayerType() gopacket.LayerType {
	return layers.LayerTypeEthernet
}

// SerializeTo serializes a data structure to byte arrays
func (d *SetGenerteIngressPortreq) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
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
	data[i] = byte(d.ECOCLength1)
	i++
	data[i] = byte(d.ECOCBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.ECOCType)
	i += 2
	data[i] = byte(d.ECOCLength2)
	i++
	binary.BigEndian.PutUint32(data[i:i+4], d.ECOCInstance)
	i += 4
	data[i] = byte(d.EndBranch)

	return nil

}

// Decode decodes byte arrays to a data structure
func (d *SetGenerteIngressPortreq) Decode(data []byte) error {
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
	d.ECOCLength1 = data[i]
	i++
	d.ECOCBranch = data[i]
	i++
	d.ECOCType = binary.BigEndian.Uint16(data[i : i+2])
	i += 2
	d.ECOCLength2 = data[i]
	i++
	d.ECOCInstance = binary.BigEndian.Uint32(data[i : i+4])
	i += 4
	d.EndBranch = data[i]

	return nil
}
