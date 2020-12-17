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
	"reflect"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGenerateDiscoverySOLICIT(t *testing.T) {
	tests := []struct {
		name string
		want gopacket.SerializableLayer
	}{
		// TODO: Add test cases.
		{
			name: "GenerateDiscoverySOLICIT-1",
			want: &DiscoverySolicit{
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateDiscoverySOLICIT(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateDiscoverySOLICIT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoverySolicit_String(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		CPType         uint8
		CPLength       uint16
		CPValue        []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "String-1",
			fields: fields{
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
			},
			want: fmt.Sprintf("Opcode:253, DiscoveryType:1, VendorType:254, Length:37, VendorIDType:253, VendorIDLength:3, VendorID:2aea15, CPType:5, CPLength:1, CPValue:80, NWType:6, NWLength:16, NWValue:%v, DVType:7, DVLength:1, DVValue:01, SCPType:8, SCPLength:1, SCPValue:03", hex.EncodeToString([]byte("tibitcom.com"))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoverySolicit{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				CPType:         tt.fields.CPType,
				CPLength:       tt.fields.CPLength,
				CPValue:        tt.fields.CPValue,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("DiscoverySolicit.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoverySolicit_Len(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		CPType         uint8
		CPLength       uint16
		CPValue        []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "Len-1",
			fields: fields{
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
			},
			want: 46,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoverySolicit{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				CPType:         tt.fields.CPType,
				CPLength:       tt.fields.CPLength,
				CPValue:        tt.fields.CPValue,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if got := d.Len(); got != tt.want {
				t.Errorf("DiscoverySolicit.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoverySolicit_LayerType(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		CPType         uint8
		CPLength       uint16
		CPValue        []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   gopacket.LayerType
	}{
		// TODO: Add test cases.
		{
			name: "LayerType-1",
			want: layers.LayerTypeEthernet,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoverySolicit{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				CPType:         tt.fields.CPType,
				CPLength:       tt.fields.CPLength,
				CPValue:        tt.fields.CPValue,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if got := d.LayerType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiscoverySolicit.LayerType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoverySolicit_SerializeTo(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		CPType         uint8
		CPLength       uint16
		CPValue        []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	type args struct {
		b    gopacket.SerializeBuffer
		opts gopacket.SerializeOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "SerializeTo-1",
			fields: fields{
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
			},
			args: args{
				b:    gopacket.NewSerializeBufferExpectedSize(0, 36),
				opts: gopacket.SerializeOptions{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoverySolicit{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				CPType:         tt.fields.CPType,
				CPLength:       tt.fields.CPLength,
				CPValue:        tt.fields.CPValue,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if err := d.SerializeTo(tt.args.b, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("DiscoverySolicit.SerializeTo error = %v, wantErr %v", err, tt.wantErr)
			}

			data := tt.args.b.Bytes()
			cnt := 0
			digits := 0
			if !reflect.DeepEqual(d.Opcode, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo Opcode error")
			}
			cnt++
			if !reflect.DeepEqual(d.DiscoveryType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo DiscoveryType error")
			}
			cnt++
			if !reflect.DeepEqual(d.VendorType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo VendorType error")
			}
			cnt++

			if !reflect.DeepEqual(d.Length, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo Length error")
			}
			cnt += 2

			if !reflect.DeepEqual(d.VendorIDType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo VendorIDType error")
			}
			cnt++
			if !reflect.DeepEqual(d.VendorIDLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo VendorIDLength error")
			}
			cnt += 2
			digits = int(d.VendorIDLength)
			if !reflect.DeepEqual(d.VendorID, data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.SerializeTo VendorID error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.CPType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo CPType error")
			}
			cnt++
			if !reflect.DeepEqual(d.CPLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo CPLength error")
			}
			cnt += 2
			digits = int(d.CPLength)
			if !reflect.DeepEqual(d.CPValue, data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.SerializeTo CPValue error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.NWType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo NWType error")
			}
			cnt++
			if !reflect.DeepEqual(d.NWLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo NWLength error")
			}
			cnt += 2
			//digits = int(d.NWLength)
			for _, dvalue := range d.NWValue {
				if !reflect.DeepEqual(dvalue, data[cnt]) {
					t.Error("DiscoverySolicit.SerializeTo NWValue error")
				}
				cnt++
			}
			// if !reflect.DeepEqual(string(d.NWValue), string(data[cnt:cnt+digits])) {
			// 	t.Errorf("DiscoverySolicit.SerializeTo NWValue error %x %x", d.NWValue, data[cnt:cnt+digits])
			// }
			cnt += 4
			if !reflect.DeepEqual(d.DVType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo DVType error")
			}
			cnt++
			if !reflect.DeepEqual(d.DVLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo DVLength error")
			}
			cnt += 2
			digits = int(d.DVLength)
			if !reflect.DeepEqual(d.DVValue, data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.SerializeTo DVValue error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.SCPType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo SCPType error")
			}
			cnt++
			if !reflect.DeepEqual(d.SCPLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo SCPLength error")
			}
			cnt += 2
			digits = int(d.SCPLength)
			if !reflect.DeepEqual(d.SCPValue, data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.SerializeTo SCPValue error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.PadType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo PadType error")
			}
			cnt++
			if !reflect.DeepEqual(d.PadLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo PadLength error")
			}
			cnt += 2
			digits = int(d.PadLength)
			if !reflect.DeepEqual(d.PadValue, data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.SerializeTo PadValue error")
			}

		})
	}
}

func TestDiscoverySolicit_Decode(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		CPType         uint8
		CPLength       uint16
		CPValue        []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Decode-1",
			args: args{
				data: []byte{0xfd, 0x01, 0xfe, 0x00, 0x25, 0xfd, 0x00, 0x03, 0x2a, 0xea, 0x15, 0x05, 0x00, 0x01, 0x80, 0x06, 0x00, 0x10, 0x74, 0x69, 0x62, 0x69, 0x74, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x00, 0x00, 0x07, 0x00, 0x01, 0x01, 0x08, 0x00, 0x01, 0x03, 0xff, 0x00, 0x01, 0x00},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoverySolicit{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				CPType:         tt.fields.CPType,
				CPLength:       tt.fields.CPLength,
				CPValue:        tt.fields.CPValue,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if err := d.Decode(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("DiscoverySolicit.Decode error = %v, wantErr %v", err, tt.wantErr)
			}

			cnt := 0
			digits := 0

			if !reflect.DeepEqual(d.Opcode, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode Opcode error Opcode=%x,data=%x,cnt=%x", d.Opcode, tt.args.data[cnt:], cnt)
			}
			cnt++
			if !reflect.DeepEqual(d.DiscoveryType, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode DiscoveryType error DiscoveryType=%X data=%x", d.DiscoveryType, tt.args.data[cnt:cnt+1])
			}
			cnt++
			if !reflect.DeepEqual(d.VendorType, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode VendorType error %x %x", tt.args.data[cnt], d.VendorType)
			}
			cnt++

			if !reflect.DeepEqual(d.Length, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Errorf("DiscoverySolicit.Decode Length error Length=%x data[cnt:cnt+2]=%x", string(d.Length), string(binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])))
			}
			cnt += 2

			if !reflect.DeepEqual(d.VendorIDType, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode VendorIDType error VendorIDType=%x data=%x", d.VendorIDType, tt.args.data[cnt])
			}
			cnt++
			if !reflect.DeepEqual(d.VendorIDLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Errorf("DiscoverySolicit.Decode VendorIDLength error VendorIDLength=%x data[cnt:cnt+2]=%x", d.VendorIDLength, tt.args.data[cnt:cnt+2])
			}
			cnt += 2
			digits = int(d.VendorIDLength)
			if !reflect.DeepEqual(d.VendorID, tt.args.data[cnt:cnt+digits]) {
				t.Errorf("DiscoverySolicit.Decode VendorID error %v %v", d.VendorID, tt.args.data[cnt:cnt+3])
			}
			cnt += digits
			if !reflect.DeepEqual(d.CPType, tt.args.data[cnt]) {
				t.Error("DiscoverySolicit.Decode CPType error")
			}
			cnt++
			if !reflect.DeepEqual(d.CPLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.Decode CPLength error")
			}
			cnt += 2
			digits = int(d.CPLength)
			if !reflect.DeepEqual(d.CPValue, tt.args.data[cnt:cnt+digits]) {
				t.Errorf("DiscoverySolicit.Decode CPValue error %x %x", tt.args.data[14:15], d.CPValue)
			}
			cnt += digits
			if !reflect.DeepEqual(d.NWType, tt.args.data[cnt]) {
				t.Error("DiscoverySolicit.Decode NWType error")
			}
			cnt++
			if !reflect.DeepEqual(d.NWLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.Decode NWLength error")
			}
			cnt += 2
			//digits = int(d.NWLength)
			for _, dvalue := range d.NWValue {
				if !reflect.DeepEqual(dvalue, tt.args.data[cnt]) {
					t.Errorf("DiscoverySolicit.Decode NWValue error %x %x", dvalue, tt.args.data[cnt])
				}
				cnt++
			}
			// if !reflect.DeepEqual(string(d.NWValue), string(tt.args.data[cnt:cnt+digits])) {
			// 	t.Errorf("DiscoverySolicit.Decode NWValue error %x %x", d.NWValue, tt.args.data[cnt:cnt+digits])
			// }
			// cnt += 4
			if !reflect.DeepEqual(d.DVType, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode DVType error %x %x", d.DVType, tt.args.data[cnt])
			}
			cnt++
			if !reflect.DeepEqual(d.DVLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Errorf("DiscoverySolicit.Decode DVLength error %x %x", d.DVLength, tt.args.data[cnt:cnt+2])
			}
			cnt += 2
			digits = int(d.DVLength)
			if !reflect.DeepEqual(d.DVValue, tt.args.data[cnt:cnt+digits]) {
				t.Errorf("DiscoverySolicit.Decode DVValue error %x %x", d.DVValue, tt.args.data[cnt:cnt+digits])
			}
			cnt += digits
			if !reflect.DeepEqual(d.SCPType, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode SCPType error %x %x", d.SCPType, tt.args.data[cnt])
			}
			cnt++
			if !reflect.DeepEqual(d.SCPLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.Decode SCPLength error")
			}
			cnt += 2
			digits = int(d.SCPLength)
			if !reflect.DeepEqual(d.SCPValue, tt.args.data[cnt:cnt+digits]) {
				t.Errorf("DiscoverySolicit.Decode SCPValue error cnt= %x digits=%x", cnt, digits)
			}
			cnt += digits
			if !reflect.DeepEqual(d.PadType, tt.args.data[cnt]) {
				t.Error("DiscoverySolicit.Decode PadType error")
			}
			cnt++
			if !reflect.DeepEqual(d.PadLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.Decode PadLength error")
			}
			cnt += 2
			digits = int(d.PadLength)
			if !reflect.DeepEqual(d.PadValue, tt.args.data[cnt:cnt+digits]) {
				t.Errorf("DiscoverySolicit.Decode PadValue error %x %x", d.PadValue, tt.args.data[cnt:cnt+digits])
			}

		})
	}
}

func TestDiscoveryHello_String(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		TunnelType     uint8
		TunnelLength   uint16
		TunnelValue    []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "String-1",
			fields: fields{
				Opcode:         0xfd,
				DiscoveryType:  0x01,
				VendorType:     0xfe,
				Length:         0x25,
				VendorIDType:   0xfd,
				VendorIDLength: 3,
				VendorID:       []byte{0x2a, 0xea, 0x15},
				NWType:         0x06,
				NWLength:       16,
				NWValue:        []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16},
				DVType:         0x07,
				DVLength:       1,
				DVValue:        []byte{2},
				SCPType:        0x08,
				SCPLength:      1,
				SCPValue:       []byte{0x03},
				TunnelType:     0x0a,
				TunnelLength:   2,
				TunnelValue:    []byte{0x00, 0x01},
				PadType:        0xff,
				PadLength:      1,
				PadValue:       []byte{0},
			},
			want: "Opcode:253, DiscoveryType:1, VendorType:254, Length:37, VendorIDType:253, VendorIDLength:3, VendorID:2aea15, NWType:6, NWLength:16, NWValue:01020304050607080910111213141516, DVType:7, DVLength:1, DVValue:02, SCPType:8, SCPLength:1, SCPValue:03, TunnelType:10, TunnelLength:2, TunnelValue:0001, PadType:255, PadLength:1, PadValue:00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoveryHello{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				TunnelType:     tt.fields.TunnelType,
				TunnelLength:   tt.fields.TunnelLength,
				TunnelValue:    tt.fields.TunnelValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("DiscoveryHello.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoveryHello_Len(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		TunnelType     uint8
		TunnelLength   uint16
		TunnelValue    []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "Len-1",
			fields: fields{
				Opcode:         0xfd,
				DiscoveryType:  0x01,
				VendorType:     0xfe,
				Length:         0x25,
				VendorIDType:   0xfd,
				VendorIDLength: 3,
				VendorID:       []byte{0x2a, 0xea, 0x15},
				NWType:         0x06,
				NWLength:       16,
				NWValue:        []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16},
				DVType:         0x07,
				DVLength:       1,
				DVValue:        []byte{2},
				SCPType:        0x08,
				SCPLength:      1,
				SCPValue:       []byte{0x03},
				TunnelType:     0x0a,
				TunnelLength:   2,
				TunnelValue:    []byte{0x00, 0x01},
				PadType:        0xff,
				PadLength:      1,
				PadValue:       []byte{0},
			},
			want: 46,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoveryHello{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				TunnelType:     tt.fields.TunnelType,
				TunnelLength:   tt.fields.TunnelLength,
				TunnelValue:    tt.fields.TunnelValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if got := d.Len(); got != tt.want {
				t.Errorf("DiscoveryHello.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoveryHello_LayerType(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		TunnelType     uint8
		TunnelLength   uint16
		TunnelValue    []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   gopacket.LayerType
	}{
		// TODO: Add test cases.
		{
			name: "LayerType-1",
			want: layers.LayerTypeEthernet,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoveryHello{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				TunnelType:     tt.fields.TunnelType,
				TunnelLength:   tt.fields.TunnelLength,
				TunnelValue:    tt.fields.TunnelValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if got := d.LayerType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiscoveryHello.LayerType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoveryHello_SerializeTo(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		TunnelType     uint8
		TunnelLength   uint16
		TunnelValue    []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	type args struct {
		b    gopacket.SerializeBuffer
		opts gopacket.SerializeOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "SerializeTo-1",
			fields: fields{
				Opcode:         0xfd,
				DiscoveryType:  0x01,
				VendorType:     0xfe,
				Length:         0x25,
				VendorIDType:   0xfd,
				VendorIDLength: 3,
				VendorID:       []byte{0x2a, 0xea, 0x15},
				NWType:         0x06,
				NWLength:       16,
				NWValue:        []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16},
				DVType:         0x07,
				DVLength:       1,
				DVValue:        []byte{2},
				SCPType:        0x08,
				SCPLength:      1,
				SCPValue:       []byte{0x03},
				TunnelType:     0x0a,
				TunnelLength:   2,
				TunnelValue:    []byte{0x00, 0x01},
				PadType:        0xff,
				PadLength:      1,
				PadValue:       []byte{0},
			},
			args: args{
				b:    gopacket.NewSerializeBufferExpectedSize(0, 46),
				opts: gopacket.SerializeOptions{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoveryHello{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				TunnelType:     tt.fields.TunnelType,
				TunnelLength:   tt.fields.TunnelLength,
				TunnelValue:    tt.fields.TunnelValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if err := d.SerializeTo(tt.args.b, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("DiscoveryHello.SerializeTo() error = %v, wantErr %v", err, tt.wantErr)
			}

			data := tt.args.b.Bytes()
			cnt := 0
			digits := 0
			if !reflect.DeepEqual(d.Opcode, data[cnt]) {
				t.Errorf("DiscoverySolicit.SerializeTo Opcode error Opcode=%x,data=%x,cnt=%x", d.Opcode, data[cnt:], cnt)
			}
			cnt++
			if !reflect.DeepEqual(d.DiscoveryType, data[cnt]) {
				t.Errorf("DiscoverySolicit.SerializeTo DiscoveryType error DiscoveryType=%X data=%x", d.DiscoveryType, data[cnt:cnt+1])
			}
			cnt++
			if !reflect.DeepEqual(d.VendorType, data[cnt]) {
				t.Errorf("DiscoverySolicit.SerializeTo VendorType error %x %x", data[cnt], d.VendorType)
			}
			cnt++

			if !reflect.DeepEqual(d.Length, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Errorf("DiscoverySolicit.SerializeTo Length error Length=%x data[cnt:cnt+2]=%x", string(d.Length), string(binary.BigEndian.Uint16(data[cnt:cnt+2])))
			}
			cnt += 2

			if !reflect.DeepEqual(d.VendorIDType, data[cnt]) {
				t.Errorf("DiscoverySolicit.SerializeTo VendorIDType error VendorIDType=%x data=%x", d.VendorIDType, data[cnt])
			}
			cnt++
			if !reflect.DeepEqual(d.VendorIDLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Errorf("DiscoverySolicit.SerializeTo VendorIDLength error VendorIDLength=%x data[cnt:cnt+2]=%x", d.VendorIDLength, data[cnt:cnt+2])
			}
			cnt += 2
			digits = int(d.VendorIDLength)
			if !reflect.DeepEqual(d.VendorID, data[cnt:cnt+digits]) {
				t.Errorf("DiscoverySolicit.SerializeTo VendorID error %v %v", d.VendorID, data[cnt:cnt+3])
			}
			cnt += digits
			if !reflect.DeepEqual(d.NWType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo NWType error")
			}
			cnt++
			if !reflect.DeepEqual(d.NWLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo NWLength error")
			}
			cnt += 2
			//digits = int(d.NWLength)
			for _, dvalue := range d.NWValue {
				if !reflect.DeepEqual(dvalue, data[cnt]) {
					t.Errorf("DiscoverySolicit.SerializeTo NWValue error %x %x", dvalue, data[cnt])
				}
				cnt++
			}
			// if !reflect.DeepEqual(string(d.NWValue), string(data[cnt:cnt+digits])) {
			// 	t.Errorf("DiscoverySolicit.SerializeTo NWValue error %x %x", d.NWValue, data[cnt:cnt+digits])
			// }
			// cnt += 16
			if !reflect.DeepEqual(d.DVType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo DVType error")
			}
			cnt++
			if !reflect.DeepEqual(d.DVLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo DVLength error")
			}
			cnt += 2
			digits = int(d.DVLength)
			if !reflect.DeepEqual(d.DVValue, data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.SerializeTo DVValue error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.SCPType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo SCPType error")
			}
			cnt++
			if !reflect.DeepEqual(d.SCPLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo SCPLength error")
			}
			cnt += 2
			digits = int(d.SCPLength)
			if !reflect.DeepEqual(d.SCPValue, data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.SerializeTo SCPValue error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.TunnelType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo TunnelType error")
			}
			cnt++
			if !reflect.DeepEqual(d.TunnelLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo TunnelLength error")
			}
			cnt += 2
			digits = int(d.TunnelLength)
			if !reflect.DeepEqual(d.TunnelValue, data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.SerializeTo TunnelValue error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.PadType, data[cnt]) {
				t.Error("DiscoverySolicit.SerializeTo PadType error")
			}
			cnt++
			if !reflect.DeepEqual(d.PadLength, binary.BigEndian.Uint16(data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.SerializeTo PadLength error")
			}
			cnt += 2
			digits = int(d.PadLength)
			if !reflect.DeepEqual(d.PadValue, data[cnt:cnt+digits]) {
				t.Errorf("DiscoverySolicit.SerializeTo PadValue error %x %x", d.PadValue, data[cnt:cnt+digits])
			}
		})
	}
}

func TestDiscoveryHello_Decode(t *testing.T) {
	type fields struct {
		BaseLayer      layers.BaseLayer
		Opcode         uint8
		DiscoveryType  uint8
		VendorType     uint8
		Length         uint16
		VendorIDType   uint8
		VendorIDLength uint16
		VendorID       []byte
		NWType         uint8
		NWLength       uint16
		NWValue        []byte
		DVType         uint8
		DVLength       uint16
		DVValue        []byte
		SCPType        uint8
		SCPLength      uint16
		SCPValue       []byte
		TunnelType     uint8
		TunnelLength   uint16
		TunnelValue    []byte
		PadType        uint8
		PadLength      uint16
		PadValue       []byte
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Decode-1",
			args: args{
				data: []byte{0xfd, 0x02, 0xfe, 0x00, 0x26, 0xfd, 0x00, 0x03, 0x2a, 0xea, 0x15, 0x06, 0x00, 0x10, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x07, 0x00, 0x01, 0x02, 0x08, 0x00, 0x01, 0x03, 0x0a, 0x00, 0x02, 0x00, 0x01, 0xff, 0x00, 0x01, 0x00},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DiscoveryHello{
				BaseLayer:      tt.fields.BaseLayer,
				Opcode:         tt.fields.Opcode,
				DiscoveryType:  tt.fields.DiscoveryType,
				VendorType:     tt.fields.VendorType,
				Length:         tt.fields.Length,
				VendorIDType:   tt.fields.VendorIDType,
				VendorIDLength: tt.fields.VendorIDLength,
				VendorID:       tt.fields.VendorID,
				NWType:         tt.fields.NWType,
				NWLength:       tt.fields.NWLength,
				NWValue:        tt.fields.NWValue,
				DVType:         tt.fields.DVType,
				DVLength:       tt.fields.DVLength,
				DVValue:        tt.fields.DVValue,
				SCPType:        tt.fields.SCPType,
				SCPLength:      tt.fields.SCPLength,
				SCPValue:       tt.fields.SCPValue,
				TunnelType:     tt.fields.TunnelType,
				TunnelLength:   tt.fields.TunnelLength,
				TunnelValue:    tt.fields.TunnelValue,
				PadType:        tt.fields.PadType,
				PadLength:      tt.fields.PadLength,
				PadValue:       tt.fields.PadValue,
			}
			if err := d.Decode(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("DiscoveryHello.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			cnt := 0
			digits := 0
			if !reflect.DeepEqual(d.Opcode, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode Opcode error Opcode=%x,data=%x,cnt=%x", d.Opcode, tt.args.data[cnt:], cnt)
			}
			cnt++
			if !reflect.DeepEqual(d.DiscoveryType, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode DiscoveryType error DiscoveryType=%X data=%x", d.DiscoveryType, tt.args.data[cnt:cnt+1])
			}
			cnt++
			if !reflect.DeepEqual(d.VendorType, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode VendorType error %x %x", tt.args.data[cnt], d.VendorType)
			}
			cnt++

			if !reflect.DeepEqual(d.Length, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Errorf("DiscoverySolicit.Decode Length error Length=%x data[cnt:cnt+2]=%x", string(d.Length), string(binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])))
			}
			cnt += 2

			if !reflect.DeepEqual(d.VendorIDType, tt.args.data[cnt]) {
				t.Errorf("DiscoverySolicit.Decode VendorIDType error VendorIDType=%x data=%x", d.VendorIDType, tt.args.data[cnt])
			}
			cnt++
			if !reflect.DeepEqual(d.VendorIDLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Errorf("DiscoverySolicit.Decode VendorIDLength error VendorIDLength=%x data[cnt:cnt+2]=%x", d.VendorIDLength, tt.args.data[cnt:cnt+2])
			}
			cnt += 2
			digits = int(d.VendorIDLength)
			if !reflect.DeepEqual(d.VendorID, tt.args.data[cnt:cnt+digits]) {
				t.Errorf("DiscoverySolicit.Decode VendorID error %v %v", d.VendorID, tt.args.data[cnt:cnt+3])
			}
			cnt += digits
			if !reflect.DeepEqual(d.NWType, tt.args.data[cnt]) {
				t.Error("DiscoverySolicit.Decode NWType error")
			}
			cnt++
			if !reflect.DeepEqual(d.NWLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.Decode NWLength error")
			}
			cnt += 2
			//digits = int(d.NWLength)
			for _, dvalue := range d.NWValue {
				if !reflect.DeepEqual(dvalue, tt.args.data[cnt]) {
					t.Errorf("DiscoverySolicit.Decode NWValue error %x %x", dvalue, tt.args.data[cnt])
				}
				cnt++
			}
			// if !reflect.DeepEqual(string(d.NWValue), string(tt.args.data[cnt:cnt+digits])) {
			// 	t.Errorf("DiscoverySolicit.Decode NWValue error %x %x", d.NWValue, tt.args.data[cnt:cnt+digits])
			// }
			// cnt += 16
			if !reflect.DeepEqual(d.DVType, tt.args.data[cnt]) {
				t.Error("DiscoverySolicit.Decode DVType error")
			}
			cnt++
			if !reflect.DeepEqual(d.DVLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.Decode DVLength error")
			}
			cnt += 2
			digits = int(d.DVLength)
			if !reflect.DeepEqual(d.DVValue, tt.args.data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.Decode DVValue error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.SCPType, tt.args.data[cnt]) {
				t.Error("DiscoverySolicit.Decode SCPType error")
			}
			cnt++
			if !reflect.DeepEqual(d.SCPLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.Decode SCPLength error")
			}
			cnt += 2
			digits = int(d.SCPLength)
			if !reflect.DeepEqual(d.SCPValue, tt.args.data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.Decode SCPValue error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.TunnelType, tt.args.data[cnt]) {
				t.Error("DiscoverySolicit.Decode TunnelType error")
			}
			cnt++
			if !reflect.DeepEqual(d.TunnelLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.Decode TunnelLength error")
			}
			cnt += 2
			digits = int(d.TunnelLength)
			if !reflect.DeepEqual(d.TunnelValue, tt.args.data[cnt:cnt+digits]) {
				t.Error("DiscoverySolicit.Decode TunnelValue error")
			}
			cnt += digits
			if !reflect.DeepEqual(d.PadType, tt.args.data[cnt]) {
				t.Error("DiscoverySolicit.Decode PadType error")
			}
			cnt++
			if !reflect.DeepEqual(d.PadLength, binary.BigEndian.Uint16(tt.args.data[cnt:cnt+2])) {
				t.Error("DiscoverySolicit.Decode PadLength error")
			}
			cnt += 2
			digits = int(d.PadLength)
			if !reflect.DeepEqual(d.PadValue, tt.args.data[cnt:cnt+digits]) {
				t.Errorf("DiscoverySolicit.Decode PadValue error %x %x", d.PadValue, tt.args.data[cnt:cnt+digits])
			}
		})
	}
}
