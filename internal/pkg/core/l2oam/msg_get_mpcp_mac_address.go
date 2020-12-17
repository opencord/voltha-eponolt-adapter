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
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// GetMpcpMacAddressReq is a structure for "MPCP/MAC Address" message
type GetMpcpMacAddressReq struct {
	layers.BaseLayer
	Opcode uint8
	Flags      uint16
	OAMPDUCode uint8
	OUId       []byte // Organizationally Unique Identifier: 2a:ea:15 (Tibit Communications)
	TOMIOpcode uint8
	OCBranch   uint8
	OCType     uint16
	OCLength   uint8
	OCInstance uint32
	VdBranch uint8
	VdLeaf   uint16
	EndBranch uint8
}

// GenerateGetMpcpMacAddress generates "MPCP/MAC Address" message
func GenerateGetMpcpMacAddress(oc *TomiObjectContext) gopacket.SerializableLayer {
	tibitData := &GetMpcpMacAddressReq{
		// IEEE 1904.2
		Opcode: 0x03,
		// OMI Protocol
		Flags:      0x0050,
		OAMPDUCode: 0xfe,
		OUId:       []byte{0x2a, 0xea, 0x15},
		// TiBiT OLT Management Interface
		TOMIOpcode: 0x01,
		// Object Context
		OCBranch:   oc.Branch,
		OCType:     oc.Type,
		OCLength:   oc.Length,
		OCInstance: oc.Instance,
		// Vd
		VdBranch: 0xcc,
		VdLeaf:   0x0008,
		// End
		EndBranch: 0,
	}
	return tibitData
}

// String returns the string expression of GetMpcpMacAddressReq
func (d *GetMpcpMacAddressReq) String() string {
	message := fmt.Sprintf("Opcode:%02x, Flags:%04x, OAMPDUCode:%02x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%02x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, OCBranch:%02x, OCType:%04x, OCLength:%02x, OCInstance:%08x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VdBranch:%02x, VdLeaf:%04x, EndBranch:%02x", message, d.VdBranch, d.VdLeaf, d.EndBranch)
	return message
}

// Len returns the length of GetModuleNumberRes
func (d *GetMpcpMacAddressReq) Len() int {
	return 21
}

// LayerType returns the ethernet type of GetMpcpMacAddressReq
func (d *GetMpcpMacAddressReq) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *GetMpcpMacAddressReq) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
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

// GetMpcpMacAddressRes is a structure for a response of "MPCP/MAC Address" message
type GetMpcpMacAddressRes struct {
	layers.BaseLayer
	Opcode uint8
	Flags      uint16
	OAMPDUCode uint8
	OUId       []byte // Organizationally Unique Identifier: 2a:ea:15 (Tibit Communications)
	TOMIOpcode uint8
	OCBranch   uint8
	OCType     uint16
	OCLength   uint8
	OCInstance uint32
	VcBranch uint8
	VcLeaf   uint16
	VcLength uint8
	EcLength uint8
	EcValue  []byte
	EndBranch uint8
}

// String returns the string expression of GetMpcpMacAddressRes
func (d *GetMpcpMacAddressRes) String() string {
	message := fmt.Sprintf("Opcode:%02x, Flags:%04x, OAMPDUCode:%02x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%02x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, OCBranch:%02x, OCType:%04x, OCLength:%02x, OCInstance:%08x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VcBranch:%02x, VcLeaf:%04x, VcLength:%02x", message, d.VcBranch, d.VcLeaf, d.VcLength)
	message = fmt.Sprintf("%s, EcLength:%02x, EcValue:%v, EndBranch:%02x", message, d.EcLength, hex.EncodeToString(d.EcValue), d.EndBranch)
	return message
}

// Len returns the length of GetMpcpMacAddressRes
func (d *GetMpcpMacAddressRes) Len() int {
	return 20 + int(d.VcLength) + 1
}

// LayerType returns the ethernet type of GetMpcpMacAddressRes
func (d *GetMpcpMacAddressRes) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// Decode decodes byte arrays to a data structure
func (d *GetMpcpMacAddressRes) Decode(data []byte) error {
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
	d.EcLength = data[i]
	i++
	d.EcValue = data[i : i+int(d.EcLength)]
	i += int(d.EcLength)
	d.EndBranch = data[i]

	return nil
}

// GetMacAddress returns a MAC address
func (d *GetMpcpMacAddressRes) GetMacAddress() string {
	var buf []string
	for _, b := range d.EcValue {
		buf = append(buf, fmt.Sprintf("%02x", b))
	}
	return strings.Join(buf, ":")
}
