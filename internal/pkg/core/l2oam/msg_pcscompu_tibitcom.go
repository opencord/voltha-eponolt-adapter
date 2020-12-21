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
)

func GeneratePcscompuTivitcom(oc *TomiObjectContext) gopacket.SerializableLayer {
	tibitData := &TOAMGetRequest{
		// IEEE 1904.2
		Opcode: 0x03,
		// OMI Protocol
		Flags:      0x0050,
		OAMPDUCode: 0xfe,
		OUId:       []byte{0x2a, 0xea, 0x15},
		// TiBiT OLT Management Interface
		TOMIOpcode: 0x81,
		// Correlation Tag
		CTBranch:   0x0c,
		CTType:     0x0c7a,
		CTLength:   4,
		CTInstance: 0x00000003,
		// Object Context
		OCBranch:   oc.Branch,
		OCType:     oc.Type,
		OCLength:   oc.Length,
		OCInstance: oc.Instance,
		// Vd
		VdBranch: 0x01,
		VdLeaf:   0x0007,
		// End
		EndBranch: 0,
	}

	return tibitData
}

type GetTrafficControlRefTableRes struct {
	ComResp TOAMGetResponse

	ECOC      []EcOcPONLinkTrafficControlReferenceTable
	EndBranch uint8
}

type EcOcPONLinkTrafficControlReferenceTable struct {
	EcOcLength   uint8
	EcOcBranch   uint8
	EcOcType     uint16
	EcOcLength2  uint8
	EcOcInstance []byte
}

func (d *GetTrafficControlRefTableRes) String() string {
	message := d.ComResp.String()
	for k, ec := range d.ECOC {
		message = message + "EC[" + string(k) + "],EcOcLength:" + string(ec.EcOcLength) + ",EcOcBranch:" + string(ec.EcOcBranch) + ",EcOcType:" + string(ec.EcOcType) + ",EcOcLength2" + string(ec.EcOcLength2) + ",EcOcInstance:" + hex.EncodeToString(ec.EcOcInstance)
	}
	message = fmt.Sprintf("%s", message)
	return message

}

func (d *GetTrafficControlRefTableRes) Decode(data []byte) error {
	d.ComResp.Decode(data)
	i := d.ComResp.Len()
	for _, ec := range d.ECOC {
		ec.EcOcLength = data[i]
		i++
		ec.EcOcBranch = data[i]
		i++
		binary.BigEndian.PutUint16(data[i:i+2], ec.EcOcType)
		i = i + 2
		ec.EcOcLength2 = data[i]
		i++
		copy(data[i:i+int(ec.EcOcLength2)], ec.EcOcInstance)
		i = i + int(ec.EcOcLength2)
	}

	d.EndBranch = data[i]
	i++

	return nil

}
