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

// L2switchingDomainRequest is a structure for a request of "L2 Switching Domain/Action Inlet entry" message
type L2switchingDomainRequest struct {
	layers.BaseLayer
	Opcode uint8
	Flags      uint16
	OAMPDUCode uint8
	OUId       []byte // Organizationally Unique Identifier: 2a:ea:15 (Tibit Communications)
	TOMIOpcode uint8
	CTBranch   uint8
	CTType     uint16
	CTLength   uint8
	CTInstance uint32
	OCBranch   uint8
	OCType     uint16
	OCLength   uint8
	OCInstance uint32
	VcBranch uint8
	VcLeaf   uint16
	VcLength uint8
	SOLength    uint8
	SOBranch    uint8
	SOType      uint16
	SOValLength uint8
	SOInstance  uint8
	TMLLength uint8
	TMLList   []TpidVid
	TOPopLength uint8
	TOPopValue  uint8
	TOSLength uint8
	TOSList   []TpidVid
	TOPushLength uint8
	TOPushList   []TpidVid
	EndBranch  uint8
	SizeMargin uint8
}

// TpidVid is a structure for TPID/VID set
type TpidVid struct {
	Length     uint8
	TpIDLength uint8
	TpIDValue  []byte
	VIdLength  uint8
	VIdValue   []byte
}

// String returns the string expression of L2switchingDomainRequest
func (d *L2switchingDomainRequest) String() string {
	message := fmt.Sprintf("Opcode:%x, Flags:%x, OAMPDUCode:%x, OUId:%v", d.Opcode, d.Flags, d.OAMPDUCode, hex.EncodeToString(d.OUId))
	message = fmt.Sprintf("%s, TOMIOpcode:%x", message, d.TOMIOpcode)
	message = fmt.Sprintf("%s, CTBranch:%x, CTType:%x, CTLength:%x, CTInstance:%x", message, d.CTBranch, d.CTType, d.CTLength, d.CTInstance)
	message = fmt.Sprintf("%s, OCBranch:%x, OCType:%x, OCLength:%x, OCInstance:%x", message, d.OCBranch, d.OCType, d.OCLength, d.OCInstance)
	message = fmt.Sprintf("%s, VcBranch:%x, VcLeaf:%x, VcLength:%x", message, d.VcBranch, d.VcLeaf, d.VcLength)
	message = fmt.Sprintf("%s, SOLength:%x, SOBranch:%x, SOType:%x, SOValLength:%x, SOInstance:%x", message, d.SOLength, d.SOBranch, d.SOType, d.SOValLength, d.SOInstance)
	message = fmt.Sprintf("%s, TMLLength:%x", message, d.TMLLength)
	for i, tagMatch := range d.TMLList {
		message = fmt.Sprintf("%s, %s", message, stringToTpidVid(&tagMatch, i))
	}
	message = fmt.Sprintf("%s, TOPopLength:%x, TOPopValue:%x", message, d.TOPopLength, d.TOPopValue)
	message = fmt.Sprintf("%s, TOSLength:%x", message, d.TOSLength)
	for i, tagOpSet := range d.TOSList {
		message = fmt.Sprintf("%s, %s", message, stringToTpidVid(&tagOpSet, i))
	}
	message = fmt.Sprintf("%s, TOPushLength:%x", message, d.TOPushLength)
	for i, tagOpPush := range d.TOPushList {
		message = fmt.Sprintf("%s, %s", message, stringToTpidVid(&tagOpPush, i))
	}
	return message
}

func stringToTpidVid(tpidVid *TpidVid, index int) string {
	message := fmt.Sprintf("EC[%v]:{Length:%x, TpIdLength:%x", index, tpidVid.Length, tpidVid.TpIDLength)
	if tpidVid.TpIDLength != 128 {
		message = fmt.Sprintf("%s, TpIdValue:%v", message, hex.EncodeToString(tpidVid.TpIDValue))
	}
	message = fmt.Sprintf("%s, VIdLength:%x", message, tpidVid.VIdLength)
	if tpidVid.VIdLength != 128 {
		message = fmt.Sprintf("%s, VIdValue:%v", message, hex.EncodeToString(tpidVid.VIdValue))
	}
	message = fmt.Sprintf("%s}", message)
	return message
}

// Len returns the length of L2switchingDomainRequest
func (d *L2switchingDomainRequest) Len() int {
	return 22 + int(d.CTLength) + int(d.OCLength) + int(d.VcLength) + int(d.SizeMargin)
}

// LayerType returns the ethernet type of L2switchingDomainRequest
func (d *L2switchingDomainRequest) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (d *L2switchingDomainRequest) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	plen := int(d.Len())
	data, err := b.PrependBytes(plen)
	if err != nil {
		return err
	}

	i := 0
	data[i] = byte(d.Opcode)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.Flags)
	i += 2
	data[i] = byte(d.OAMPDUCode)
	i++
	copy(data[i:i+len(d.OUId)], d.OUId)
	i += len(d.OUId)
	data[i] = byte(d.TOMIOpcode)
	i++
	data[i] = byte(d.CTBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.CTType)
	i += 2
	data[i] = byte(d.CTLength)
	i++
	binary.BigEndian.PutUint32(data[i:i+4], d.CTInstance)
	i += 4
	data[i] = byte(d.OCBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.OCType)
	i += 2
	data[i] = byte(d.OCLength)
	i++
	binary.BigEndian.PutUint32(data[i:i+4], d.OCInstance)
	i += 4
	data[i] = byte(d.VcBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.VcLeaf)
	i += 2
	data[i] = byte(d.VcLength)
	i++
	data[i] = byte(d.SOLength)
	i++
	data[i] = byte(d.SOBranch)
	i++
	binary.BigEndian.PutUint16(data[i:i+2], d.SOType)
	i += 2
	data[i] = byte(d.SOValLength)
	i++
	data[i] = byte(d.SOInstance)
	i++
	data[i] = byte(d.TMLLength)
	i++
	for _, tagMatch := range d.TMLList {
		nextIndex := serializeTpidVid(&tagMatch, data, i)
		i = nextIndex
	}
	data[i] = byte(d.TOPopLength)
	i++
	data[i] = byte(d.TOPopValue)
	i++
	data[i] = byte(d.TOSLength)
	i++
	for _, tagOpSet := range d.TOSList {
		nextIndex := serializeTpidVid(&tagOpSet, data, i)
		i = nextIndex
	}
	data[i] = d.TOPushLength
	i++
	for _, tagOpPush := range d.TOPushList {
		nextIndex := serializeTpidVid(&tagOpPush, data, i)
		i = nextIndex
	}
	data[i] = byte(d.EndBranch)

	return nil
}

func serializeTpidVid(tpidVid *TpidVid, data []byte, startIndex int) int {
	i := startIndex
	data[i] = tpidVid.Length
	i++
	ln := tpidVid.TpIDLength
	data[i] = ln
	i++
	if ln != 128 { // !Empty?
		copy(data[i:i+int(ln)], tpidVid.TpIDValue)
		i += int(ln)
	}
	ln = tpidVid.VIdLength
	data[i] = ln
	i++
	if ln != 128 { // !Empty?
		copy(data[i:i+int(ln)], tpidVid.VIdValue)
		i += int(ln)
	}
	return i
}
