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

// OampduTlvTypeLocalInfo means that TLV type is "Local Information"
var OampduTlvTypeLocalInfo uint8 = 0x01

// OampduTlvTypeRemoteInfo means that TLV type is "Remote Information"
var OampduTlvTypeRemoteInfo uint8 = 0x02

// OampduTlvTypeOrganizationSpecificInfo means that TLV type is "Organization Specific Information"
var OampduTlvTypeOrganizationSpecificInfo uint8 = 0xfe

// GeneateKeepAlive1 generates "OAMPDU Information(first)"
func GeneateKeepAlive1(isOnuPkgA bool) gopacket.SerializableLayer {
	var osiLength uint8
	var osiValue []byte
	if isOnuPkgA {
		osiLength = 7
		osiValue = []byte{0x00, 0x10, 0x00, 0x00, 0x23}
	} else {
		osiLength = 8
		osiValue = []byte{0x90, 0x82, 0x60, 0x02, 0x01, 0x01}
	}
	tibitData := &OAMPDUInformation{
		// IEEE 1904.2
		Opcode: 0x03,
		// OAM Protocol
		Flags: 0x0008,
		Code:  0x00,
		// Local Information TLV
		LIType:   OampduTlvTypeLocalInfo,
		LILength: 16,
		LIValue:  []byte{0x01, 0x00, 0x00, 0x00, 0x1b, 0x04, 0xb0, 0x2a, 0xea, 0x15, 0x00, 0x00, 0x00, 0x23},
		// Organization Specific Information TLV
		OSIType:   OampduTlvTypeOrganizationSpecificInfo,
		OSILength: osiLength,
		OSIValue:  osiValue,
	}

	return tibitData
}

// GeneateKeepAlive2 generates "OAMPDU Information(second)"
func GeneateKeepAlive2(riValue []byte, isOnuPkgA bool) gopacket.SerializableLayer {
	var osiLength uint8
	var osiValue []byte
	if isOnuPkgA {
		osiLength = 7
		osiValue = []byte{0x00, 0x10, 0x00, 0x00, 0x23}
	} else {
		osiLength = 8
		osiValue = []byte{0x90, 0x82, 0x60, 0x02, 0x01, 0x01}
	}
	tibitData := &OAMPDUInformation{
		// IEEE 1904.2
		Opcode: 0x03,
		// OAM Protocol
		Flags: 0x0030,
		Code:  0x00,
		// Local Information TLV
		LIType:   OampduTlvTypeLocalInfo,
		LILength: 16,
		LIValue:  []byte{0x01, 0x00, 0x00, 0x00, 0x1b, 0x04, 0xb0, 0x2a, 0xea, 0x15, 0x00, 0x00, 0x00, 0x23},
		// Remote Information TLV
		RIType:   OampduTlvTypeRemoteInfo,
		RILength: 16,
		RIValue:  riValue,
		// Organization Specific Information TLV
		OSIType:   OampduTlvTypeOrganizationSpecificInfo,
		OSILength: osiLength,
		OSIValue:  osiValue,
	}

	return tibitData

}

// GeneateKeepAlive3 generates "OAMPDU Information(third)"
func GeneateKeepAlive3(riValue []byte) gopacket.SerializableLayer {
	tibitData := &OAMPDUInformation{
		// IEEE 1904.2
		Opcode: 0x03,
		// OAM Protocol
		Flags: 0x0050,
		Code:  0x00,
		// Local Information TLV
		LIType:   OampduTlvTypeLocalInfo,
		LILength: 16,
		LIValue:  []byte{0x01, 0x00, 0x00, 0x00, 0x1b, 0x04, 0xb0, 0x2a, 0xea, 0x15, 0x00, 0x00, 0x00, 0x23},
		// Remote Information TLV
		RIType:   OampduTlvTypeRemoteInfo,
		RILength: 16,
		RIValue:  riValue,
	}

	return tibitData

}

// OAMPDUInformation is a structure for "OAMPDU Information" message
type OAMPDUInformation struct {
	layers.BaseLayer
	Opcode uint8
	Flags uint16
	Code  uint8
	LIType   uint8
	LILength uint8
	LIValue  []byte
	RIType   uint8
	RILength uint8
	RIValue  []byte
	OSIType   uint8
	OSILength uint8
	OSIValue  []byte
}

// String returns the string expression of OAMPDUInformation
func (d *OAMPDUInformation) String() string {
	message := fmt.Sprintf("Opcode:%v, Flags:%v, Code:%v", d.Opcode, d.Flags, d.Code)
	message = fmt.Sprintf("%s, LIType:%v, LILength:%v, LIValue:%v", message, d.LIType, d.LILength, hex.EncodeToString(d.LIValue))
	message = fmt.Sprintf("%s, RIType:%v, RILength:%v, RIValue:%v", message, d.RIType, d.RILength, hex.EncodeToString(d.RIValue))
	message = fmt.Sprintf("%s, OSIType:%v, OSILength:%v, OSIValue:%v", message, d.OSIType, d.OSILength, hex.EncodeToString(d.OSIValue))
	return message
}

// Len returns the length of OAMPDUInformation
func (d *OAMPDUInformation) Len() int {
	len := (1) + (3) + int(d.LILength) + int(d.RILength) + int(d.OSILength)
	return len
}

// LayerType returns the ethernet type of OAMPDUInformation
func (d *OAMPDUInformation) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *OAMPDUInformation) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	plen := int(d.Len())
	data, err := b.PrependBytes(plen)
	if err != nil {
		return err
	}

	i := 0
	data[i] = byte(d.Opcode)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.Flags)
	i = i + 2
	data[i] = byte(d.Code)
	i++

	if d.LILength != 0 {
		data[i] = byte(d.LIType)
		data[i+1] = byte(d.LILength)
		copy(data[i+2:i+int(d.LILength)], d.LIValue)
		i = i + int(d.LILength)
	}

	if d.RILength != 0 {
		data[i] = byte(d.RIType)
		data[i+1] = byte(d.RILength)
		copy(data[i+2:i+int(d.RILength)], d.RIValue)
		i = i + int(d.RILength)
	}

	if d.OSILength != 0 {
		data[i] = byte(d.OSIType)
		data[i+1] = byte(d.OSILength)
		copy(data[i+2:i+int(d.OSILength)], d.OSIValue)
		//i = i + int(d.OSILength)
	}

	return nil
}

// Decode decodes byte arrays to a data structure
func (d *OAMPDUInformation) Decode(data []byte) error {
	i := 0
	d.Opcode = data[i]
	i++

	d.Flags = binary.BigEndian.Uint16(data[i : i+2])
	i = i + 2
	d.Code = data[i]
	i++

	for {
		if len(data) <= i {
			break
		}
		tlvType := data[i]
		tlvLength := data[i+1]
		if tlvLength == 0 {
			break
		}
		tlvValue := data[i+2 : i+int(tlvLength)]
		i = i + int(tlvLength)

		switch tlvType {
		case OampduTlvTypeLocalInfo:
			d.LIType = tlvType
			d.LILength = tlvLength
			d.LIValue = tlvValue
		case OampduTlvTypeRemoteInfo:
			d.RIType = tlvType
			d.RILength = tlvLength
			d.RIValue = tlvValue
		case OampduTlvTypeOrganizationSpecificInfo:
			d.OSIType = tlvType
			d.OSILength = tlvLength
			d.OSIValue = tlvValue
		default:
			return fmt.Errorf("tlvType Error: %v", tlvType)
		}
	}

	return nil
}

// IsOnuPkgA returns true if message type of OAMPDUInformation is PkgA
func (d *OAMPDUInformation) IsOnuPkgA() bool {
	if d.OSILength == 0 {
		// return true if OSILength is 0, this means the message is KeepAlive3
		return true
	} else if d.OSILength == 7 {
		return (d.OSIValue[0] == 0x00 && d.OSIValue[1] == 0x10 && d.OSIValue[2] == 0x00)
	}
	return false
}

// IsOnuPkgB returns true if message type of OAMPDUInformation is PkgA
func (d *OAMPDUInformation) IsOnuPkgB() bool {
	if d.OSILength == 0 {
		// return true if OSILength is 0, this means the message is KeepAlive3
		return true
	} else if d.OSILength == 7 {
		return (d.OSIValue[0] == 0x90 && d.OSIValue[1] == 0x82 && d.OSIValue[2] == 0x60)
	}
	return false
}
