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
	"github.com/google/gopacket"
)

// GenerateOnuSystemInfo generates "ONU System Information" message
func GenerateOnuSystemInfo(PkgType string) gopacket.SerializableLayer {

	if PkgType == OnuPkgTypeA {
		//TypeA
		data := &TibitFrame{
			Data: []byte{
				0x03, 0x00, 0x50, 0xfe, 0x00, 0x10, 0x00, 0x01,
				0xd7, 0x00, 0x06, 0x00, 0x00,
			},
		}

		return data
	} else if PkgType == OnuPkgTypeB {
		//TypeB
		data := &TibitFrame{
			Data: []byte{
				0x03, 0x00, 0x50, 0xfe, 0x90, 0x82, 0x60, 0x00,
				0xca, 0xfe, 0x00, 0xb7, 0x00, 0x40, 0x00, 0x00,
			},
		}

		return data
	}

	return nil
}

// GetOnuSerialNo returns a serial number
func GetOnuSerialNo(pkgType string, data []byte) string {
	if pkgType == OnuPkgTypeA {
		// 030050fe00100002 00 - 07
		// d70006404d493920 08 - 15
		// 32304b2042303041 16 - 23
		// 3730204130302031 24 - 31
		// 3730382020202020
		// 2000000030313233
		// 3435363730313233
		// 3435363738394041
		// 42434445
		return string(data[20:26])
	}
	return ""
}

// GetOnuManufacture returns a manufacture information
func GetOnuManufacture(pkgType string, data []byte) string {
	if pkgType == OnuPkgTypeA {
		if data[16] == 0x32 && data[17] == 0x30 && data[18] == 0x4b {
			return "FURUKAWA"
		}
	}
	return ""
}
