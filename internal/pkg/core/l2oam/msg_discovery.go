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

// GenerateDiscoverySOLICIT generates "Discovery: SOLICIT" message
func GenerateDiscoverySOLICIT() gopacket.SerializableLayer {

	tibitData := &DiscoverySolicit{
		// IEEE 1904.2
		Opcode:        0xfd,
		DiscoveryType: 0x01,
		// Vendor-specific
		VendorType: 0xfe,
		Length:     37,
		// Vendor ID
		VendorIDType:   0xfd,
		VendorIDLength: 3,
		VendorID:       []byte{0x2a, 0xea, 0x15},
		// Controller Priority
		CPType:   0x05,
		CPLength: 1,
		CPValue:  []byte{128},
		//NetworkID
		NWType:   0x06,
		NWLength: 16,
		NWValue:  []byte("tibitcom.com"),
		//Device Type
		DVType:   0x07,
		DVLength: 1,
		DVValue:  []byte{1},
		//Supported CLient Protocols
		SCPType:   0x08,
		SCPLength: 1,
		SCPValue:  []byte{0x03},
		//Padding
		PadType:   0xff,
		PadLength: 1,
		PadValue:  []byte{0},
	}

	return tibitData
}

// DiscoverySolicit is a structure for Discovery message
type DiscoverySolicit struct {
	layers.BaseLayer
	Opcode        uint8
	DiscoveryType uint8
	VendorType uint8
	Length     uint16 // length of after this data without Padding
	VendorIDType   uint8
	VendorIDLength uint16
	VendorID       []byte
	CPType   uint8
	CPLength uint16
	CPValue  []byte
	NWType   uint8
	NWLength uint16
	NWValue  []byte
	DVType   uint8
	DVLength uint16
	DVValue  []byte
	SCPType   uint8
	SCPLength uint16
	SCPValue  []byte
	PadType   uint8
	PadLength uint16
	PadValue  []byte
}

// String returns the string expression of DiscoverySolicit
func (d *DiscoverySolicit) String() string {
	message := fmt.Sprintf("Opcode:%v, DiscoveryType:%v, VendorType:%v, Length:%v", d.Opcode, d.DiscoveryType, d.VendorType, d.Length)
	message = fmt.Sprintf("%s, VendorIDType:%v, VendorIDLength:%v, VendorID:%v", message, d.VendorIDType, d.VendorIDLength, hex.EncodeToString(d.VendorID))
	message = fmt.Sprintf("%s, CPType:%v, CPLength:%v, CPValue:%v", message, d.CPType, d.CPLength, hex.EncodeToString(d.CPValue))
	message = fmt.Sprintf("%s, NWType:%v, NWLength:%v, NWValue:%v", message, d.NWType, d.NWLength, hex.EncodeToString(d.NWValue))
	message = fmt.Sprintf("%s, DVType:%v, DVLength:%v, DVValue:%v", message, d.DVType, d.DVLength, hex.EncodeToString(d.DVValue))
	message = fmt.Sprintf("%s, SCPType:%v, SCPLength:%v, SCPValue:%v", message, d.SCPType, d.SCPLength, hex.EncodeToString(d.SCPValue))
	return message
}

// Len returns the length of DiscoverySolicit
func (d *DiscoverySolicit) Len() int {
	len := (2) + (3) + int(d.Length) + (int(d.PadLength) + 3)
	return len
}

// LayerType returns the ethernet type of DiscoverySolicit
func (d *DiscoverySolicit) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *DiscoverySolicit) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	plen := int(d.Len())
	data, err := b.PrependBytes(plen)
	if err != nil {
		return err
	}

	i := 0
	data[i] = byte(d.Opcode)
	i++
	data[i] = byte(d.DiscoveryType)
	i++
	data[i] = byte(d.VendorType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.Length)
	i = i + 2

	data[i] = byte(d.VendorIDType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.VendorIDLength)
	i = i + 2
	copy(data[i:i+int(d.VendorIDLength)], d.VendorID)
	i = i + int(d.VendorIDLength)

	data[i] = byte(d.CPType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.CPLength)
	i = i + 2
	copy(data[i:i+int(d.CPLength)], d.CPValue)
	i = i + int(d.CPLength)

	data[i] = byte(d.NWType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.NWLength)
	i = i + 2
	copy(data[i:i+int(d.NWLength)], d.NWValue)
	i = i + int(d.NWLength)

	data[i] = byte(d.DVType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.DVLength)
	i = i + 2
	copy(data[i:i+int(d.DVLength)], d.DVValue)
	i = i + int(d.DVLength)

	data[i] = byte(d.SCPType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.SCPLength)
	i = i + 2
	copy(data[i:i+int(d.SCPLength)], d.SCPValue)
	i = i + int(d.SCPLength)

	data[i] = byte(d.PadType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.PadLength)
	i = i + 2
	copy(data[i:i+int(d.PadLength)], d.PadValue)

	return nil
}

// Decode decodes byte arrays to a data structure
func (d *DiscoverySolicit) Decode(data []byte) error {
	i := 0
	d.Opcode = data[i]
	i++
	d.DiscoveryType = data[i]
	i++
	d.VendorType = data[i]
	i++
	d.Length = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2

	d.VendorIDType = data[i]
	i++
	d.VendorIDLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.VendorID = data[i : i+int(d.VendorIDLength)]
	i = i + int(d.VendorIDLength)

	d.CPType = data[i]
	i++
	d.CPLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.CPValue = data[i : i+int(d.CPLength)]
	i = i + int(d.CPLength)

	d.NWType = data[i]
	i++
	d.NWLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.NWValue = data[i : i+int(d.NWLength)]
	i = i + int(d.NWLength)

	d.DVType = data[i]
	i++
	d.DVLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.DVValue = data[i : i+int(d.DVLength)]
	i = i + int(d.DVLength)

	d.SCPType = data[i]
	i++
	d.SCPLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.SCPValue = data[i : i+int(d.SCPLength)]
	i = i + int(d.SCPLength)

	d.PadType = data[i]
	i++
	d.PadLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.PadValue = data[i : i+int(d.PadLength)]

	return nil
}

// DiscoveryHello is a structure for Discovery message
type DiscoveryHello struct {
	layers.BaseLayer
	Opcode        uint8
	DiscoveryType uint8
	VendorType uint8
	Length     uint16 // length of after this data without Padding
	VendorIDType   uint8
	VendorIDLength uint16
	VendorID       []byte
	NWType   uint8
	NWLength uint16
	NWValue  []byte
	DVType   uint8
	DVLength uint16
	DVValue  []byte
	SCPType   uint8
	SCPLength uint16
	SCPValue  []byte
	TunnelType   uint8
	TunnelLength uint16
	TunnelValue  []byte
	PadType   uint8
	PadLength uint16
	PadValue  []byte
}

// String returns the string expression of DiscoveryHello
func (d *DiscoveryHello) String() string {
	message := fmt.Sprintf("Opcode:%v, DiscoveryType:%v, VendorType:%v, Length:%v", d.Opcode, d.DiscoveryType, d.VendorType, d.Length)
	message = fmt.Sprintf("%s, VendorIDType:%v, VendorIDLength:%v, VendorID:%v", message, d.VendorIDType, d.VendorIDLength, hex.EncodeToString(d.VendorID))
	message = fmt.Sprintf("%s, NWType:%v, NWLength:%v, NWValue:%v", message, d.NWType, d.NWLength, hex.EncodeToString(d.NWValue))
	message = fmt.Sprintf("%s, DVType:%v, DVLength:%v, DVValue:%v", message, d.DVType, d.DVLength, hex.EncodeToString(d.DVValue))
	message = fmt.Sprintf("%s, SCPType:%v, SCPLength:%v, SCPValue:%v", message, d.SCPType, d.SCPLength, hex.EncodeToString(d.SCPValue))
	message = fmt.Sprintf("%s, TunnelType:%v, TunnelLength:%v, TunnelValue:%v", message, d.TunnelType, d.TunnelLength, hex.EncodeToString(d.TunnelValue))
	message = fmt.Sprintf("%s, PadType:%v, PadLength:%v, PadValue:%v", message, d.PadType, d.PadLength, hex.EncodeToString(d.PadValue))
	return message
}

// Len returns the length of DiscoveryHello
func (d *DiscoveryHello) Len() int {
	len := (2) + (3) + int(d.Length) + (int(d.PadLength) + 3)
	return len
}

// LayerType returns the ethernet type of DiscoveryHello
func (d *DiscoveryHello) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *DiscoveryHello) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	plen := int(d.Len())
	data, err := b.PrependBytes(plen)
	if err != nil {
		return err
	}

	i := 0
	data[i] = byte(d.Opcode)
	i++
	data[i] = byte(d.DiscoveryType)
	i++
	data[i] = byte(d.VendorType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.Length)
	i = i + 2

	data[i] = byte(d.VendorIDType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.VendorIDLength)
	i = i + 2
	copy(data[i:i+int(d.VendorIDLength)], d.VendorID)
	i = i + int(d.VendorIDLength)

	data[i] = byte(d.NWType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.NWLength)
	i = i + 2
	copy(data[i:i+int(d.NWLength)], d.NWValue)
	i = i + int(d.NWLength)

	data[i] = byte(d.DVType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.DVLength)
	i = i + 2
	copy(data[i:i+int(d.DVLength)], d.DVValue)
	i = i + int(d.DVLength)

	data[i] = byte(d.SCPType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.SCPLength)
	i = i + 2
	copy(data[i:i+int(d.SCPLength)], d.SCPValue)
	i = i + int(d.SCPLength)

	data[i] = byte(d.TunnelType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.TunnelLength)
	i = i + 2
	copy(data[i:i+int(d.TunnelLength)], d.TunnelValue)
	i = i + int(d.TunnelLength)

	data[i] = byte(d.PadType)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.PadLength)
	i = i + 2
	copy(data[i:i+int(d.PadLength)], d.PadValue)

	return nil
}

// Decode decodes byte arrays to a data structure
func (d *DiscoveryHello) Decode(data []byte) error {
	i := 0
	d.Opcode = data[i]
	i++
	d.DiscoveryType = data[i]
	i++
	d.VendorType = data[i]
	i++
	d.Length = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2

	d.VendorIDType = data[i]
	i++
	d.VendorIDLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.VendorID = data[i : i+int(d.VendorIDLength)]
	i = i + int(d.VendorIDLength)

	d.NWType = data[i]
	i++
	d.NWLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.NWValue = data[i : i+int(d.NWLength)]
	i = i + int(d.NWLength)

	d.DVType = data[i]
	i++
	d.DVLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.DVValue = data[i : i+int(d.DVLength)]
	i = i + int(d.DVLength)

	d.SCPType = data[i]
	i++
	d.SCPLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.SCPValue = data[i : i+int(d.SCPLength)]
	i = i + int(d.SCPLength)

	d.TunnelType = data[i]
	i++
	d.TunnelLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.TunnelValue = data[i : i+int(d.TunnelLength)]
	i = i + int(d.TunnelLength)

	d.PadType = data[i]
	i++
	d.PadLength = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.PadValue = data[i : i+int(d.PadLength)]

	return nil
}
