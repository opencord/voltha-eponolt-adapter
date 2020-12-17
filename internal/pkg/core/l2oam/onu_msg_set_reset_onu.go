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

// GenerateSetResetOnu generates reset ONU message
func GenerateSetResetOnu(pkgType string) gopacket.SerializableLayer {
	if pkgType == OnuPkgTypeB {
		return &TibitFrame{
			Data: []byte{
				0x03, 0x00, 0x50, 0xfe, 0x90, 0x82, 0x60, 0x03,
				0xca, 0xfe, 0x00, 0xb9, 0x00, 0x0e, 0x01, 0x01,
				0x00, 0x00,
			},
		}
	}
	return &TibitFrame{
		Data: []byte{
			0x03, 0x00, 0x50, 0xfe, 0x00, 0x10, 0x00, 0x03,
			0xd6, 0x00, 0x00, 0xd9, 0x00, 0x01, 0x80, 0x00,
			0x00,
		},
	}

}
