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
package core

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	oop "github.com/opencord/voltha-protos/v3/go/openolt"
)

const vlanEnable = true
const vlanIdManagement = 4090

var svlanTagAuth = []byte{0x88, 0xa8, 0x0f, 0xfb}

// IfName is a Interface Name of this device
var IfName string

// SrcMac is a MAC address of this device
var SrcMac string

var l2oamHandle *L2oamHandle

// GetL2oamHandle returns my handle
func GetL2oamHandle() *L2oamHandle {
	return l2oamHandle
}

// L2oamHandle contains handle information
type L2oamHandle struct {
	Handle *pcap.Handle
	ReqCh  chan []byte
}

// NewL2oamHandle L2oamHandle created
func NewL2oamHandle(ctx context.Context, ifname string, srcMac string) {
	logger.Debug(ctx, fmt.Sprintf("L2oamHandle start. ifname=%s, srcMac=%s", ifname, srcMac))
	IfName = ifname
	SrcMac = srcMac
	if GetL2oamHandle() == nil {
		handle, _ := pcap.OpenLive(IfName, int32(1500), false, time.Millisecond*1000)
		l2oamHandle = &L2oamHandle{
			Handle: handle,
			ReqCh:  make(chan []byte),
		}
		if handle == nil {
			logger.Debug(ctx, "could not get pcap handle.")
		} else {
			go l2oamHandle.startThread(context.Background())
		}
	}
}

// because of each handle object that is provided for a device receive messages respectively,
// call this method using goroutine
func (h *L2oamHandle) startThread(ctx context.Context) {
	logger.Debug(ctx, "L2oamHandle thread start. ")

	var filter string = "ether proto 0xa8c8 or 0x888e or 0x8100 or 0x88a8"
	err := h.Handle.SetBPFFilter(filter)
	if err != nil {
		logger.Error(ctx, "Error: Handle.SetBPFFilter")
	}

	packetSource := gopacket.NewPacketSource(h.Handle, h.Handle.LinkType())

	for {
		select {
		// send a message directly using send method, not sending it via channels
		case message := <-h.ReqCh:
			logger.Debug(ctx, fmt.Sprintf("gopacket send. %x", message))
			if err := h.Handle.WritePacketData(message); err != nil {
				logger.Debug(ctx, "write-packet-error")
			}
		// Packets are received
		case packet := <-packetSource.Packets():
			etherLayer := packet.Layer(layers.LayerTypeEthernet)
			logger.Debug(ctx, fmt.Sprintf("gopacket receive. %x", etherLayer))
			if etherLayer != nil {
				// type assertion
				etherPacket, _ := etherLayer.(*layers.Ethernet)
				packetBytes := append(etherPacket.Contents, etherPacket.Payload...)
				logger.Debug(ctx, fmt.Sprintf("gopacket ether receive. %x", packetBytes))
				if vlanEnable {
					h.dispatchVlan(ctx, etherPacket)
				} else {
					h.dispatch(ctx, etherPacket)
				}
			}
		}
	}
}
func (h *L2oamHandle) dispatch(ctx context.Context, etherPacket *layers.Ethernet) {
	logger.Debug(ctx, fmt.Sprintf("dispatch(). %x", etherPacket))

	etherType := uint16(etherPacket.EthernetType)
	if etherType == uint16(layers.EthernetTypeEAPOL) {
		logger.Debug(ctx, fmt.Sprintf("receive message = EAPOL, Contents=%x, Payload=%x", etherPacket.Contents, etherPacket.Payload))
		device := FindL2oamDevice(hex.EncodeToString(etherPacket.SrcMAC))
		if device == nil {
			// unregisterd device
			logger.Error(ctx, fmt.Sprintf("Received from an unregistered device. macAddr=%s", etherPacket.SrcMAC))
			h.packetIn(etherPacket)
		} else {
			device.receiveEapol(etherPacket)
		}

	} else if etherType == 0xa8c8 {
		bytes := etherPacket.Payload
		opcode := bytes[0]
		oampduCode := bytes[3]
		logger.Debug(ctx, fmt.Sprintf("receive message = 1904.2 OAM opcode=%v oampduCode=%v,  %x", opcode, oampduCode, bytes))

		device := FindL2oamDevice(hex.EncodeToString(etherPacket.SrcMAC))
		if device == nil {
			// unregisterd device
			logger.Error(ctx, fmt.Sprintf("Received from an unregistered device. macAddr=%s", etherPacket.SrcMAC))
			return
		}

		// OAM: keep-alive message
		if opcode == 0x03 && oampduCode == 0x00 {
			device.recieveKeepAlive(etherPacket)

		} else {
			device.recieve(etherPacket)
		}
	} else if etherType == uint16(layers.EthernetTypeDot1Q) {
		device := FindL2oamDevice(hex.EncodeToString(etherPacket.SrcMAC))
		if device == nil {
			// unregisterd device
			logger.Error(ctx, fmt.Sprintf("Received from an unregistered device. macAddr=%s", etherPacket.SrcMAC))
			h.packetIn(etherPacket)
		} else {
			device.receiveEapol(etherPacket)
		}

	} else {
		logger.Error(ctx, fmt.Sprintf("unknown etherType = 0x%X", etherType))
	}
}
func (h *L2oamHandle) dispatchVlan(ctx context.Context, etherPacket *layers.Ethernet) {
	logger.Debug(ctx, fmt.Sprintf("dispatchVlan(). %x", etherPacket))

	etherType := uint16(etherPacket.EthernetType)
	if etherType == uint16(layers.EthernetTypeQinQ) {
		// Remove the VLAN tag
		vlanPacket := &layers.Dot1Q{}
		if err := vlanPacket.DecodeFromBytes(etherPacket.Payload, nil); err != nil {
			logger.Debug(ctx, fmt.Sprintf("error:%v", err))
		}
		etherBytes := etherPacket.Contents
		binary.BigEndian.PutUint16(etherBytes[12:], uint16(vlanPacket.Type))
		packetBytes := append(etherBytes, vlanPacket.Payload...)
		etherPacket = &layers.Ethernet{}
		if err := etherPacket.DecodeFromBytes(packetBytes, nil); err != nil {
			logger.Debug(ctx, fmt.Sprintf("error:%v", err))
		}
		h.dispatch(ctx, etherPacket)
	} else {
		logger.Error(ctx, fmt.Sprintf("Discard unsupported etherType. 0x%X", etherType))
	}
}

func (h *L2oamHandle) send(message []byte) {
	if h.Handle != nil {
		h.ReqCh <- message
	}
}
func (h *L2oamHandle) packetIn(etherPacket *layers.Ethernet) {
	packetBytes := append(etherPacket.Contents, etherPacket.Payload...)
	pktInd := &oop.PacketIndication{
		IntfId:    0,
		GemportId: 1,
		PortNo:    16,
		Pkt:       packetBytes,
		IntfType:  "pon",
	}
	logger.Info(context.Background(), fmt.Sprintf("L2oamHandle.packetIn() pktInd=%x", pktInd))

	if err := l2oamDeviceHandler.handlePacketIndication(context.Background(), pktInd); err != nil {
		logger.Error(context.Background(), "L2oamHandle.packetIn() error. ")
	}
}
