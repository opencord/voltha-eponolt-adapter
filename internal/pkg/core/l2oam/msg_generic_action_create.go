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
	"encoding/hex"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	ActionTypeProtocolFilter = 1

	ActionTypeTrafficProfile = 2
)

// GenerateGenericActionCreate generates "Generic/Action Create" message
func GenerateGenericActionCreate(actionType int) gopacket.SerializableLayer {
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
		OCInstance: 00000000,
		//Vc
		VcBranch: 0x6e,
		VcLeaf:   0x7001,
		VcLength: 4,
		//OT
		OT: []OTGenericActionCreate{{OTLength: 3, OTValue: getObjectType(actionType)}},
		// End
		EndBranch: 0,
	}

	return tibitData
}

func getObjectType(actionType int) []byte {
	switch actionType {
	case ActionTypeProtocolFilter:
		return []byte{0x0c, 0x0c, 0xff}
	case ActionTypeTrafficProfile:
		return []byte{0x0c, 0x07, 0x0f}
	default:
		return []byte{0x00, 0x00, 0x00}
	}
}

type setGenericActionCreateReq struct {
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
	OT []OTGenericActionCreate
	EndBranch uint8
}

// OTGenericActionCreate is a structure for "Object Type"
type OTGenericActionCreate struct {
	OTLength uint8
	OTValue  []byte
}

// String returns the string expression of setGenericActionCreateReq
func (d *setGenericActionCreateReq) String() string {
	message := fmt.Sprintf("Opcode:%x, Flags:%x, OAMPDUCode:%x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, CTBranch:%x, CTType:%x, CTLength:%x, CTInstance:%x", message, d.CTBranch, d.CTType, d.CTLength, d.CTInstance)
	message = fmt.Sprintf("%s, OCBranch:%x, OCType:%x, OCLength:%x, OCInstance:%x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VcBranch:%x, VcLeaf:%x, VcLength:%x", message, d.VcBranch, d.VcLeaf, d.VcLength)
	for _, ot := range d.OT {
		message = fmt.Sprintf("%s, OTLength:%x, OTValue:%v", message, ot.OTLength, hex.EncodeToString(ot.OTValue))
	}
	message = fmt.Sprintf("%s, EndBranch:%x", message, d.EndBranch)
	return message
}

// Len returns the length of setGenericActionCreateReq
func (d *setGenericActionCreateReq) Len() int {
	return 21 + int(d.CTLength) + int(d.OCLength) + int(d.VcLength)

}

// LayerType returns the ethernet type of setGenericActionCreateReq
func (d *setGenericActionCreateReq) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *setGenericActionCreateReq) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
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
	for _, ot := range d.OT {
		nextIndex := serializeObjectType(&ot, data, i)
		i = nextIndex
	}

	data[i] = byte(d.EndBranch)

	return nil

}

func serializeObjectType(ot *OTGenericActionCreate, data []byte, startIndex int) int {
	i := startIndex
	data[i] = ot.OTLength
	i++
	copy(data[i:i+int(ot.OTLength)], ot.OTValue)
	i += int(ot.OTLength)

	return i

}

// SetGenericActionCreateRes is a structure for a response of "Generic/Action Create"
type SetGenericActionCreateRes struct {
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

	ObOCLength   uint8
	ObOCBranch   uint8
	ObOCType     uint16
	ObOCLength2  uint8
	ObOCInstance []byte

	EndBranch uint8
}

// ObOC is a structure for "Object OC"
type ObOC struct {
	ObOCLength   uint8
	ObOCBranch   uint8
	ObOCType     uint16
	ObOCLength2  uint8
	ObOCInstance []byte
}

// String returns the string expression of SetGenericActionCreateRes
func (d *SetGenericActionCreateRes) String() string {
	message := fmt.Sprintf("Opcode:%x, Flags:%x, OAMPDUCode:%x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, CTBranch:%x, CTType:%x, CTLength:%x, CTInstance:%x", message, d.CTBranch, d.CTType, d.CTLength, d.CTInstance)
	message = fmt.Sprintf("%s, OCBranch:%x, OCType:%x, OCLength:%x, OCInstance:%x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VcBranch:%x, VcLeaf:%x, VcLength:%x", message, d.VcBranch, d.VcLeaf, d.VcLength)
	message = fmt.Sprintf("%s, ObOCLength:%x, ObOCBranch:%x, ObOCType:%x, ObOCLength2:%x, ObOCInstance:%x, EndBranch:%x",
		message, d.ObOCLength, d.ObOCBranch, d.ObOCType, d.ObOCLength2, d.ObOCInstance, d.EndBranch)
	return message
}

// Decode decodes byte arrays to a data structure
func (d *SetGenericActionCreateRes) Decode(data []byte) error {
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

	d.ObOCLength = data[i]
	i++
	d.ObOCBranch = data[i]
	i++
	d.ObOCType = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.ObOCLength2 = data[i]
	i++
	d.ObOCInstance = data[i : i+int(d.ObOCLength2)]
	i = i + int(d.ObOCLength2)

	d.EndBranch = data[i]

	return nil
}

// GetTrafficProfile returns the traffic profile of this message
func (d *SetGenericActionCreateRes) GetTrafficProfile() []byte {
	return d.ObOCInstance
}
