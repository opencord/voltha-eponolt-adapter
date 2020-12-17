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

	"github.com/google/gopacket"
)

// GenerateSetActionReset generates "Device/Action Reset" message
func GenerateSetActionReset() gopacket.SerializableLayer {

	data := &TibitFrame{
		Data: []byte{
			0x03, 0x00, 0x50, 0xfe, 0x2a, 0xea, 0x15, 0x03,
			0x0c, 0x0c, 0x7a, 0x04, 0x00, 0x00, 0x00, 0x00,
			0x0c, 0x0d, 0xce, 0x04, 0x00, 0x00, 0x00, 0x00,
			0xde, 0x70, 0x01, 0x80, 0x00, 0x00},
	}

	binary.BigEndian.PutUint32(data.Data[12:16], getOltInstance())

	return data

}
