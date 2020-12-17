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

// GenerateGetFirmwareVersion generates "Running Firmware Version" message
func GenerateGetFirmwareVersion() gopacket.SerializableLayer {
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
		VdLeaf:   0x001b,
		// End
		EndBranch: 0,
	}
	return tibitData
}

// GetFirmwareVersionRes is a structure for a response of "Running Firmware Version"
type GetFirmwareVersionRes struct {
	ComResp   TOAMGetResponse
	AKLength  uint8
	AKValue   []byte
	RVLength  uint8
	RVValue   []byte
	BSLength  uint8
	BSValue   []byte
	RNLength  uint8
	RNValue   []byte
	BDLength  uint8
	BDValue   []byte
	EndBranch uint8
}

// String returns the string expression of GetFirmwareVersionRes
func (d *GetFirmwareVersionRes) String() string {
	message := d.ComResp.String()
	message = fmt.Sprintf("%s, AKLength:%02x, AKValue:%s, RVLength:%02x, RVValue:%s", message, d.AKLength, hex.EncodeToString(d.AKValue), d.RVLength, hex.EncodeToString(d.RVValue))
	message = fmt.Sprintf("%s, BSLength:%02x, BSValue:%s, RNLength:%02x, RNValue:%s", message, d.BSLength, hex.EncodeToString(d.BSValue), d.RNLength, hex.EncodeToString(d.RNValue))
	message = fmt.Sprintf("%s, BDLength:%02x, BDValue:%s, EndBranch:%02x", message, d.BDLength, hex.EncodeToString(d.BDValue), d.EndBranch)
	return message
}

// Len returns the length of GetFirmwareVersionRes
func (d *GetFirmwareVersionRes) Len() int {
	return d.ComResp.Len() + int(d.ComResp.VcLength) + 1
}

// Decode decodes byte arrays to a data structure
func (d *GetFirmwareVersionRes) Decode(data []byte) error {
	d.ComResp.Decode(data)
	i := d.ComResp.Len()
	d.AKLength = data[i]
	i++
	d.AKValue = data[i : i+int(d.AKLength)]
	i += int(d.AKLength)
	d.RVLength = data[i]
	i++
	d.RVValue = data[i : i+int(d.RVLength)]
	i += int(d.RVLength)
	d.BSLength = data[i]
	i++
	d.BSValue = data[i : i+int(d.BSLength)]
	i += int(d.BSLength)
	d.RNLength = data[i]
	i++
	d.RNValue = data[i : i+int(d.RNLength)]
	i += int(d.RNLength)
	d.BDLength = data[i]
	i++
	d.BDValue = data[i : i+int(d.BDLength)]
	i += int(d.BDLength)
	d.EndBranch = data[i]

	return nil
}

// GetFirmwareVersionNumber returns a firmware version number
func (d *GetFirmwareVersionRes) GetFirmwareVersionNumber() string {
	return string(d.RVValue)
}
