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
	"net"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// OnuPkgType describes target package type
var OnuPkgType = "PkgB"

// OnuPkgTypeA is a constant of package
const OnuPkgTypeA = "PkgA"

// OnuPkgTypeB is a constant of package
const OnuPkgTypeB = "PkgB"

// TibitFrame is a typical structure of a frame
type TibitFrame struct {
	layers.BaseLayer
	Data []byte
}

// LayerType returns ethernet layer type
func (t *TibitFrame) LayerType() gopacket.LayerType { return layers.LayerTypeEthernet }

// SerializeTo serializes a data structure to byte arrays
func (t *TibitFrame) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	length := len(t.Data)
	bytes, err := b.PrependBytes(length)
	if err != nil {
		return err
	}
	copy(bytes, t.Data)
	return nil
}

// TomiObjectContext is a structure for tomi object
type TomiObjectContext struct {
	Branch   uint8
	Type     uint16
	Length   uint8
	Instance uint32
}

// CreateMessage creates l2 message
func CreateMessage(srcMac string, dstMac string, ethernetType layers.EthernetType, tibitData gopacket.SerializableLayer) []byte {
	srcMAC, _ := net.ParseMAC(srcMac)
	dstMAC, _ := net.ParseMAC(dstMac)

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       srcMAC,
		DstMAC:       dstMAC,
		EthernetType: ethernetType,
	}

	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(
		buf,
		gopacket.SerializeOptions{
			ComputeChecksums: true,
			FixLengths:       true,
		},
		ethernetLayer,
		tibitData,
	); err != nil {
		return buf.Bytes()
	}
	return buf.Bytes()
}

// CreateMessage creates vlan message
func CreateMessageVlan(srcMac string, dstMac string, ethernetType layers.EthernetType, tibitData gopacket.SerializableLayer, vlanLayer *layers.Dot1Q) []byte {
	srcMAC, _ := net.ParseMAC(srcMac)
	dstMAC, _ := net.ParseMAC(dstMac)

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       srcMAC,
		DstMAC:       dstMAC,
		EthernetType: ethernetType,
	}

	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(
		buf,
		gopacket.SerializeOptions{
			ComputeChecksums: true,
			FixLengths:       true,
		},
		ethernetLayer,
		vlanLayer,
		tibitData,
	); err != nil {
		return buf.Bytes()
	}
	return buf.Bytes()
}

// CorrelationTag instance ID
// It is incremented automatically for each TOMI message with OLT
var instance uint32 = 0x5c1f6a60
var instanceMutex sync.Mutex

func getOltInstance() uint32 {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()
	instance = instance + 1

	return instance

}
