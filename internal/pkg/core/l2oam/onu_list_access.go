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
	"encoding/json"
	"io/ioutil"
	"os"
)

// JSON file name
const jsonname = "/onu_list.json"

// OnuStatus has ONU status
type OnuStatus struct {
	ID           string `json:"id"`
	AdminState   string `json:"admin_state"`
	OpeState     string `json:"ope_state"`
	ConnectState string `json:"con_state"`
	MacAddress   string `json:"mac_addr"`
	RebootState  string `json:"reboot_state"`
}

// ReadOnuStatusList reads JSON file
func ReadOnuStatusList() ([]OnuStatus, error) {
	bytes, err := ioutil.ReadFile(os.Getenv("HOME") + jsonname)
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	var onuList []OnuStatus
	if err := json.Unmarshal(bytes, &onuList); err != nil {
		return nil, err
	}
	return onuList, nil
}

// WriteOnuStatusList writes JSON file
func WriteOnuStatusList(list []OnuStatus) error {
	bytes, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(os.Getenv("HOME")+jsonname, bytes, 0644)
}

// AddOnu adds ONU to ONU status list
func AddOnu(sts *OnuStatus) error {
	list, err := ReadOnuStatusList()
	if err != nil {
		return err
	}
	if list == nil {
		newList := []OnuStatus{*sts}
		return WriteOnuStatusList(newList)
	}
	return WriteOnuStatusList(append(list, *sts))
}

// UpdateOnu updates ONU status
func UpdateOnu(upSts *OnuStatus) error {
	list, err := ReadOnuStatusList()
	if (err != nil) || (list == nil) {
		return err
	}
	newList := []OnuStatus{}
	for _, sts := range list {
		if sts.ID == upSts.ID {
			newList = append(newList, *upSts)
		} else {
			newList = append(newList, sts)
		}
	}
	return WriteOnuStatusList(newList)
}

// RemoveOnu removes ONU from ONU status list
func RemoveOnu(id string) error {
	list, err := ReadOnuStatusList()
	if (err != nil) || (list == nil) {
		return err
	}
	newList := []OnuStatus{}
	for _, sts := range list {
		if sts.ID != id {
			newList = append(newList, sts)
		}
	}
	return WriteOnuStatusList(newList)
}

// GetOnuFromDeviceID returns ONU status from ONU status list using its device ID
func GetOnuFromDeviceID(id string) (*OnuStatus, error) {
	list, err := ReadOnuStatusList()
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, nil
	}
	for _, sts := range list {
		if sts.ID != id {
			return &sts, nil
		}
	}
	return nil, nil
}

// GetOnuFromMacAddr returns ONU status from ONU status list using its MAC address
func GetOnuFromMacAddr(addr string) (*OnuStatus, error) {
	list, err := ReadOnuStatusList()
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, nil
	}
	for _, sts := range list {
		if sts.MacAddress == addr {
			return &sts, nil
		}
	}
	return nil, nil
}
