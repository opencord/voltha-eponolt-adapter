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

// TOAMGetRequest is a structure for GET request of TOAM message
type TOAMGetRequest struct {
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
	VdBranch uint8
	VdLeaf   uint16
	EndBranch uint8
}

// String returns the string expression of TOAMGetRequest
func (d *TOAMGetRequest) String() string {
	message := fmt.Sprintf("Opcode:%02x, Flags:%04x, OAMPDUCode:%02x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%02x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, CTBranch:%02x, CTType:%04x, CTLength:%02x, CTInstance:%08x", message, d.CTBranch, d.CTType, d.CTLength, d.CTInstance)
	message = fmt.Sprintf("%s, OCBranch:%02x, OCType:%04x, OCLength:%02x, OCInstance:%08x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VdBranch:%02x, VdLeaf:%04x, EndBranch:%02x", message, d.VdBranch, d.VdLeaf, d.EndBranch)
	return message
}

// Len returns the length of TOAMGetRequest
func (d *TOAMGetRequest) Len() int {
	return 29
}

// LayerType returns the ethernet type of TOAMGetRequest
func (d *TOAMGetRequest) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *TOAMGetRequest) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
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
	data[i] = byte(d.VdBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.VdLeaf)
	i += 2
	data[i] = byte(d.EndBranch)

	return nil
}

// TOAMGetResponse is a structure for GET response of TOAM message
type TOAMGetResponse struct {
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
}

// String returns the string expression of TOAMGetResponse
func (d *TOAMGetResponse) String() string {
	message := fmt.Sprintf("Opcode:%02x, Flags:%04x, OAMPDUCode:%02x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%02x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, CTBranch:%02x, CTType:%04x, CTLength:%02x, CTInstance:%08x", message, d.CTBranch, d.CTType, d.CTLength, d.CTInstance)
	message = fmt.Sprintf("%s, OCBranch:%02x, OCType:%04x, OCLength:%02x, OCInstance:%08x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VcBranch:%02x, VcLeaf:%04x, VcLength:%02x", message, d.VcBranch, d.VcLeaf, d.VcLength)
	return message
}

// Len returns the length of TOAMGetResponse
func (d *TOAMGetResponse) Len() int {
	return 28
}

// LayerType returns the ethernet type of TOAMGetResponse
func (d *TOAMGetResponse) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// Decode decodes byte arrays to a data structure
func (d *TOAMGetResponse) Decode(data []byte) {
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
}
