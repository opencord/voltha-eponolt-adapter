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

// GenerateTrafficControlTrafficProfile generates "Traffic Control/Traffic Profile" message
func GenerateTrafficControlTrafficProfile(trafficControl []byte, trafficProfile []byte) gopacket.SerializableLayer {

	tibitData := &setTrafficControlTrafficProfileReq{
		// IEEE 1904.2
		Opcode: 0x03,
		// OMI Protocol
		Flags:      0x0050,
		OAMPDUCode: 0xfe,
		OUId:       []byte{0x2a, 0xea, 0x15}, // Organizationally Unique Identifier: 2a:ea:15 (Tibit Communications)
		// TiBiT OLT Management Interface
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
		VcLength: 0x09,

		//EC OC
		ECOC: []ECOCTrafficControlTrafficProfile{{EcOcLength: 8, EcOcBranch: 0x0c, EcOcType: 0x070f, EcOcLength2: 4, EcOcInstance: trafficProfile}},

		// End
		EndBranch: 0x00,
	}

	return tibitData
}

type setTrafficControlTrafficProfileReq struct {
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
	OCInstance []byte
	VcBranch uint8
	VcLeaf   uint16
	VcLength uint8

	ECOC []ECOCTrafficControlTrafficProfile

	EndBranch uint8
}

// ECOCTrafficControlTrafficProfile is a structure for Vc-EC OC object
type ECOCTrafficControlTrafficProfile struct {
	EcOcLength   uint8
	EcOcBranch   uint8
	EcOcType     uint16
	EcOcLength2  uint8
	EcOcInstance []byte
}

// String returns the string expression of setTrafficControlTrafficProfileReq
func (d *setTrafficControlTrafficProfileReq) String() string {
	message := fmt.Sprintf("Opcode:%x, Flags:%x, OAMPDUCode:%x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, CTBranch:%x, CTType:%x, CTLength:%x, CTInstance:%x", message, d.CTBranch, d.CTType, d.CTLength, d.CTInstance)
	message = fmt.Sprintf("%s, OCBranch:%x, OCType:%x, OCLength:%x, OCInstance:%x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VcBranch:%x, VcLeaf:%x, VcLength:%x", message, d.VcBranch, d.VcLeaf, d.VcLength)
	for _, ecoc := range d.ECOC {
		message = fmt.Sprintf("%s, EcOcLength:%x, EcOcBranch:%v, EcOcType:%x, EcOcLength2:%x, EcOcInstance:%x, ", message, ecoc.EcOcLength, ecoc.EcOcBranch, ecoc.EcOcType, ecoc.EcOcLength2, hex.EncodeToString(ecoc.EcOcInstance))
	}
	message = fmt.Sprintf("%s, EndBranch:%x", message, d.EndBranch)
	return message
}

// Len returns the length of setTrafficControlTrafficProfileReq
func (d *setTrafficControlTrafficProfileReq) Len() int {
	return 21 + int(d.CTLength) + int(d.OCLength) + int(d.VcLength)
}

// LayerType returns the ethernet type of setTrafficControlTrafficProfileReq
func (d *setTrafficControlTrafficProfileReq) LayerType() gopacket.LayerType {
	return layers.LayerTypeEthernet
}

// SerializeTo serializes a data structure to byte arrays
func (d *setTrafficControlTrafficProfileReq) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
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
	copy(data[i:i+4], d.OCInstance)
	i += 4
	data[i] = byte(d.VcBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.VcLeaf)
	i += 2
	data[i] = byte(d.VcLength)
	i++
	for _, ecoc := range d.ECOC {
		nextIndex := SerializeECOC(&ecoc, data, i)
		i = nextIndex
	}

	data[i] = byte(d.EndBranch)

	return nil
}

// SerializeECOC serializes a "EC OC" structure to byte arrays
func SerializeECOC(ecoc *ECOCTrafficControlTrafficProfile, data []byte, startIndex int) int {
	i := startIndex
	data[i] = byte(ecoc.EcOcLength)
	i++
	data[i] = byte(ecoc.EcOcBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], ecoc.EcOcType)
	i += 2
	data[i] = byte(ecoc.EcOcLength2)
	i++
	copy(data[i:i+int(ecoc.EcOcLength2)], ecoc.EcOcInstance)
	i += int(ecoc.EcOcLength2)

	return i
}
