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
	"github.com/google/gopacket/layers"
)

// AutonomousEvent is a structure for Autonomous Event
type AutonomousEvent struct {
	ComResp TOAMGetResponse
	EcLength  uint8
	EcValue   []byte
	EndBranch uint8
}

// String returns the string expression of AutonomousEvent
func (d *AutonomousEvent) String() string {
	message := d.ComResp.String()
	message = fmt.Sprintf("%s, EcLength:%v, EcValue:%v, EndBranch:%v", message, d.EcLength, hex.EncodeToString(d.EcValue), d.EndBranch)
	return message
}

// Len returns the length of AutonomousEvent
func (d *AutonomousEvent) Len() int {
	return d.ComResp.Len() + int(d.ComResp.VcLength) + 1
}

// LayerType returns the ethernet type of AutonomousEvent
func (d *AutonomousEvent) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *AutonomousEvent) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	return nil
}

// Decode decodes byte arrays to a data structure
func (d *AutonomousEvent) Decode(data []byte) error {
	d.ComResp.Decode(data)
	i := d.ComResp.Len()
	d.EcLength = data[i]
	i++
	d.EcValue = data[i : i+int(d.EcLength)]
	i = i + int(d.EcLength)
	d.EndBranch = data[i]

	return nil
}

// IsRegistrationStatusMessage returns true if the message is for ONU registration
func (d *AutonomousEvent) IsRegistrationStatusMessage() bool {
	return d.ComResp.VcBranch == 0x01 && d.ComResp.VcLeaf == 0x0009
}
