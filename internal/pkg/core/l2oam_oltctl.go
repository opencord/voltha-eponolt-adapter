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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type AddFlowParam struct {
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
	Value3 string `json:"value3"`
}
type AddFlowDownParam struct {
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
	Value3 string `json:"value3"`
}
type AddFlowUpParam struct {
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
	Value3 string `json:"value3"`
}

type OltCtlJson struct {
	Command          string           `json:"command"`
	Command_help     []string         `json:"command_help"`
	AddFlowParam     AddFlowParam     `json:"addFlowParam"`
	AddFlowDownParam AddFlowDownParam `json:"addFlowDownParam"`
	AddFlowUpParam   AddFlowUpParam   `json:"addFlowUpParam"`
}

func init() {
}

func oltctl_main() {
	for {
		time.Sleep(1 * time.Second)
		jsonData := readJson()
		if jsonData == nil {
			continue
		}

		switch jsonData.Command {
		case "AddFlow":
			oltctl_AddFlow(&jsonData.AddFlowParam)
		case "AddFlowDown":
			oltctl_AddFlowDown(&jsonData.AddFlowDownParam)
		case "AddFlowUp":
			oltctl_AddFlowUp(&jsonData.AddFlowUpParam)
		default:

		}
		jsonData.Command = "None"
		writeJson(jsonData)

	}

}
func oltctl_AddFlow(param *AddFlowParam) {
	logger.Debug(context.Background(), fmt.Sprintf("oltctl_AddFlow() %v", param))
}
func oltctl_AddFlowDown(param *AddFlowDownParam) {
	logger.Debug(context.Background(), fmt.Sprintf("oltctl_AddFlowDown() %v", param))
}
func oltctl_AddFlowUp(param *AddFlowUpParam) {
	logger.Debug(context.Background(), fmt.Sprintf("oltctl_AddFlowUp() %v", param))
}

func readJson() *OltCtlJson {
	bytes, err := ioutil.ReadFile(os.Getenv("HOME") + "/oltctl.json")
	if err != nil {
		fmt.Printf("json read error. %v\n", err)
		return nil
	}
	jsonData := &OltCtlJson{}
	if err := json.Unmarshal(bytes, jsonData); err != nil {
		fmt.Printf("json unmarshal error. %v\n", err)
		return nil
	}
	return jsonData
}
func writeJson(jsonData *OltCtlJson) {

	bytes, err := json.Marshal(jsonData)
	if err != nil {
		return
	}
	ioutil.WriteFile(os.Getenv("HOME")+"/oltctl.json", bytes, 0644)
}
