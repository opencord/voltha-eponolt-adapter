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

// GenerateSetVlanTagFilter generates VlanTagFilter message
func GenerateSetVlanTagFilter(pkgType string, idx int, value []byte) gopacket.SerializableLayer {

	if pkgType == OnuPkgTypeB {
		switch idx {
		case 1:
			tibitData := &TibitFrame{
				Data: []byte{
					0x03, 0x00, 0x50, 0xfe, 0x90, 0x82, 0x60, 0x03,
					0xca, 0xfe, 0x00, 0xb7, 0x00, 0x13, 0x03, 0x03,
				}, //,0x81,0x00
			}

			tibitData.Data = append(tibitData.Data, value...)

			return tibitData
		case 2:
			return &TibitFrame{
				Data: []byte{
					0x03, 0x00, 0x50, 0xfe, 0x90, 0x82, 0x60, 0x03,
					0xca, 0xfe, 0x00, 0xb7, 0x00, 0x14, 0x03, 0x01,
					0x00, 0x64,
				},
			}
		default:
			return &TibitFrame{
				Data: []byte{},
			}
		}
	} else {
		switch idx {
		case 1:
			return &TibitFrame{
				Data: []byte{
					0x03, 0x00, 0x50, 0xFE, 0x00, 0x10, 0x00, 0x03,
					0xD6, 0x00, 0x00, 0xD7, 0x05, 0x03, 0x03, 0x81,
					0x00, 0x01,
				},
			}
		case 2:
			return &TibitFrame{
				Data: []byte{
					0x03, 0x00, 0x50, 0xFE, 0x00, 0x10, 0x00, 0x03,
					0xD6, 0x00, 0x00, 0xD7, 0x05, 0x04, 0x03, 0x88,
					0xa8, 0x00,
				},
			}
		default:
			return &TibitFrame{
				Data: []byte{},
			}
		}
	}
}
