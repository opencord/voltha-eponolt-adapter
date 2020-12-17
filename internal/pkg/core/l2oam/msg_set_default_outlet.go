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

// GenerateSetDefaultOutlet generates "Default Outlet" message
func GenerateSetDefaultOutlet(oc *TomiObjectContext, onuID *TomiObjectContext) gopacket.SerializableLayer {

	tibitData := &DefaultOutletRequest{
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
		OCBranch:   oc.Branch,
		OCType:     oc.Type,
		OCLength:   oc.Length,
		OCInstance: oc.Instance,
		// Vc
		VcBranch: 0x5d,
		VcLeaf:   0x0003,
		VcLength: 0x09,
		// Default Outlet
		DOLength:    8,
		DOBranch:    0x0c,
		DOType:      0x0011,
		DOValLength: 4,
		DOInstance:  onuID.Instance,
		// End
		EndBranch: 0x00,
	}

	return tibitData
}

// DefaultOutletRequest is a structure for a request of "Default Outlet" message
type DefaultOutletRequest struct {
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
	DOLength    uint8
	DOBranch    uint8
	DOType      uint16
	DOValLength uint8
	DOInstance  uint32
	EndBranch uint8
}

// String returns the string expression of DefaultOutletRequest
func (d *DefaultOutletRequest) String() string {
	message := fmt.Sprintf("Opcode:%x, Flags:%x, OAMPDUCode:%x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, CTBranch:%x, CTType:%x, CTLength:%x, CTInstance:%x", message, d.CTBranch, d.CTType, d.CTLength, d.CTInstance)
	message = fmt.Sprintf("%s, OCBranch:%x, OCType:%x, OCLength:%x, OCInstance:%x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VcBranch:%x, VcLeaf:%x, VcLength:%x", message, d.VcBranch, d.VcLeaf, d.VcLength)
	message = fmt.Sprintf("%s, DOLength:%x, DOBranch:%x, DOType:%x, DOValLength:%x, DOInstance:%x", message, d.DOLength, d.DOBranch, d.DOType, d.DOValLength, d.DOInstance)

	return message
}

// Len returns the length of DefaultOutletRequest
func (d *DefaultOutletRequest) Len() int {
	return 38
}

// LayerType returns the ethernet type of DefaultOutletRequest
func (d *DefaultOutletRequest) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *DefaultOutletRequest) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
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
	data[i] = byte(d.DOLength)
	i++
	data[i] = byte(d.DOBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.DOType)
	i += 2
	data[i] = byte(d.DOValLength)
	i++
	binary.BigEndian.PutUint32(data[i:i+4], d.DOInstance)
	i += 4
	data[i] = byte(d.EndBranch)

	return nil
}
