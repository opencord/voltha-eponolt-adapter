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
	"strings"

	"github.com/google/gopacket"
)

// GenerateGetMacAddress generates "PON Port/MAC Address" message
func GenerateGetMacAddress() gopacket.SerializableLayer {
	tibitData := &TOAMGetRequest{
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
		OCBranch:   0x0c,
		OCType:     0x0007,
		OCLength:   4,
		OCInstance: 0x00000000,
		// Vd
		VdBranch: 0x07,
		VdLeaf:   0x0004,
		// End
		EndBranch: 0x00,
	}
	return tibitData
}

// GetMacAddressRes is a structure for a toam response
type GetMacAddressRes struct {
	ComResp TOAMGetResponse
	EcLength  uint8
	EcValue   []byte
	EndBranch uint8
}

// String returns the string expression of GetMacAddressRes
func (d *GetMacAddressRes) String() string {
	message := d.ComResp.String()
	message = fmt.Sprintf("%s, EcLength:%x, EcValue:%v, EndBranch:%x", message, d.EcLength, hex.EncodeToString(d.EcValue), d.EndBranch)
	return message
}

// Len returns the length of GetMacAddressRes
func (d *GetMacAddressRes) Len() int {
	return d.ComResp.Len() + int(d.ComResp.VcLength) + 1
}

// Decode decodes byte arrays to a data structure
func (d *GetMacAddressRes) Decode(data []byte) error {
	d.ComResp.Decode(data)
	i := d.ComResp.Len()
	d.EcLength = data[i]
	i++
	d.EcValue = data[i : i+int(d.EcLength)]
	i = i + int(d.EcLength)
	d.EndBranch = data[i]

	return nil
}

// GetMacAddress returns a MAC address
func (d *GetMacAddressRes) GetMacAddress() string {
	var buf []string
	for _, b := range d.EcValue {
		buf = append(buf, fmt.Sprintf("%02x", b))
	}
	return strings.Join(buf, ":")
}
