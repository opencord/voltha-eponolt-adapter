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

// TOAMSetRequest is a structure for SET request of TOAM message
type TOAMSetRequest struct {
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
	ECValue  []byte
	ECList []ECValueSet
	EndBranch uint8
}

// ECValueSet is a structure for Vc-EC object
type ECValueSet struct {
	Length uint8
	Value  []byte
}

// String returns the string expression of TOAMSetRequest
func (d *TOAMSetRequest) String() string {
	message := fmt.Sprintf("Opcode:%x, Flags:%x, OAMPDUCode:%x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, CTBranch:%x, CTType:%x, CTLength:%x, CTInstance:%x", message, d.CTBranch, d.CTType, d.CTLength, d.CTInstance)
	message = fmt.Sprintf("%s, OCBranch:%x, OCType:%x, OCLength:%x, OCInstance:%x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VcBranch:%x, VcLeaf:%x, VcLength:%x", message, d.VcBranch, d.VcLeaf, d.VcLength)
	message = fmt.Sprintf("%s, ECLength:%x, ECValue:%v, EndBranch:%x", message, d.ECLength, hex.EncodeToString(d.ECValue), d.EndBranch)
	return message
}

// Len returns the length of TOAMSetRequest
func (d *TOAMSetRequest) Len() int {
	if d.VcLength != 0x80 {
		return 22 + int(d.CTLength) + int(d.OCLength) + int(d.VcLength)
	}
	return 22 + int(d.CTLength) + int(d.OCLength)

}

// LayerType returns the ethernet type of TOAMSetRequest
func (d *TOAMSetRequest) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *TOAMSetRequest) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
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
	if d.VcLength != 0x80 {
		// EC
		data[i] = byte(d.ECLength)
		i++
		copy(data[i:i+int(d.ECLength)], d.ECValue)
		i += int(d.ECLength)
	}
	if d.ECList != nil {
		for _, ecSet := range d.ECList {
			data[i] = byte(ecSet.Length)
			i++
			copy(data[i:i+int(ecSet.Length)], ecSet.Value)
			i += int(ecSet.Length)
		}
	}

	data[i] = byte(d.EndBranch)

	return nil
}

// TOAMSetResponse is a structure for SET response of TOAM message
type TOAMSetResponse struct {
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
	VcBranch      uint8
	VcLeaf        uint16
	VcTOMIResCode uint8
	EndBranch uint8
}

// String returns the string expression of TOAMSetResponse
func (d *TOAMSetResponse) String() string {
	message := fmt.Sprintf("Opcode:%x, Flags:%x, OAMPDUCode:%x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, CTBranch:%x, CTType:%x, CTLength:%x, CTInstance:%x", message, d.CTBranch, d.CTType, d.CTLength, d.CTInstance)
	message = fmt.Sprintf("%s, OCBranch:%x, OCType:%x, OCLength:%x, OCInstance:%x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VcBranch:%x, VcLeaf:%x, VcTOMIResCode:%x, EndBranch:%x", message, d.VcBranch, d.VcLeaf, d.VcTOMIResCode, d.EndBranch)
	return message
}

// Len returns the length of TOAMSetResponse
func (d *TOAMSetResponse) Len() int {
	return 34 + int(d.CTLength) + int(d.OCLength)
}

// LayerType returns the ethernet type of TOAMSetResponse
func (d *TOAMSetResponse) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// Decode decodes byte arrays to a data structure
func (d *TOAMSetResponse) Decode(data []byte) error {
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
	d.VcTOMIResCode = data[i]
	i++
	d.EndBranch = data[i]

	return nil
}
