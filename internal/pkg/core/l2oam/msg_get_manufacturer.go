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

// GenerateManuFacturerInfo generates "Device/Manufacturer Info" message
func GenerateManuFacturerInfo() gopacket.SerializableLayer {

	tibitData := &TOAMGetRequest{
		// IEEE 1904.2
		Opcode: 0x03,
		//OAMPrtocol
		Flags:      0x0050,
		OAMPDUCode: 0xfe,
		OUId:       []byte{0x2a, 0xea, 0x15},
		//TiBitT OLT Management
		TOMIOpcode: 0x01,
		//Correlation Tag
		CTBranch:   0x0c,
		CTType:     0x0c7a,
		CTLength:   4,
		CTInstance: getOltInstance(),
		//Object Context
		OCBranch:   0x0c,
		OCType:     0x0dce,
		OCLength:   4,
		OCInstance: 0x00000000,
		//Vd
		VdBranch: 0xde,
		VdLeaf:   0x0006,
		//End
		EndBranch: 0x00,
	}
	return tibitData

}

// GetManufacturerRes is a structure for a toam response
type GetManufacturerRes struct {
	ComResp TOAMGetResponse
	ECLength uint8
	ECValue  []byte
	EndBranch uint8
}

// String returns the string expression of GetManufacturerRes
func (d *GetManufacturerRes) String() string {
	message := d.ComResp.String()
	message = fmt.Sprintf("%s, ECLength:%02x, ECValue:%v,EndBranch:%02x", message, d.ECLength, hex.EncodeToString(d.ECValue), d.EndBranch)
	return message
}

// Len returns the length of GetManufacturerRes
func (d *GetManufacturerRes) Len() int {
	return d.ComResp.Len() + int(d.ComResp.VcLength) + 1
}

// Decode decodes byte arrays to a data structure
func (d *GetManufacturerRes) Decode(data []byte) error {
	d.ComResp.Decode(data)
	i := d.ComResp.Len()
	d.ECLength = data[i]
	i++
	d.ECValue = data[i : i+int(d.ECLength)]
	i = i + int(d.ECLength)
	d.EndBranch = data[i]

	return nil
}

// GetManufacturer returns a manufacturer info.
func (d *GetManufacturerRes) GetManufacturer() string {
	return string(d.ECValue)
}
