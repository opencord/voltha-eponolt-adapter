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
	"encoding/hex"
	"fmt"

	"github.com/google/gopacket"
)

// GenerateGetSerialNumber generates "Device/Vendor Serial Number" message
func GenerateGetSerialNumber() gopacket.SerializableLayer {
	tibitData := &TOAMGetRequest{
		// IEEE 1904.2
		Opcode: 0x03,
		// OMI Protocol
		Flags:      0x0050,
		OAMPDUCode: 0xfe,
		OUId:       []byte{0x2a, 0xea, 0x15},
		// TiBiT OLT Management Interface
		TOMIOpcode: 0x01,
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
		// Vd
		VdBranch: 0xde,
		VdLeaf:   0x0012,
		// End
		EndBranch: 0,
	}
	return tibitData
}

// GetSerialNumberRes is a structure for toam message
type GetSerialNumberRes struct {
	ComResp   TOAMGetResponse
	EcLength  uint8
	EcValue   []byte
	EndBranch uint8
}

// String returns the string expression of GetSerialNumberRes
func (d *GetSerialNumberRes) String() string {
	message := d.ComResp.String()
	message = fmt.Sprintf("%s, EcLength:%02x, EcValue:%s EndBranch:%02x", message, d.EcLength, hex.EncodeToString(d.EcValue), d.EndBranch)
	return message
}

// Len returns the length of GetSerialNumberRes
func (d *GetSerialNumberRes) Len() int {
	return d.ComResp.Len() + int(d.ComResp.VcLength) + 1
}

// Decode decodes byte arrays to a data structure
func (d *GetSerialNumberRes) Decode(data []byte) error {
	d.ComResp.Decode(data)
	i := d.ComResp.Len()
	d.EcLength = data[i]
	i++
	d.EcValue = data[i : i+int(d.EcLength)]
	i += int(d.EcLength)
	d.EndBranch = data[i]

	return nil
}

// GetSerialNumber returns a serial number
func (d *GetSerialNumberRes) GetSerialNumber() string {
	return string(d.EcValue)
}
