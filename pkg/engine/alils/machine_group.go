package alils

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

// MachineGroupAttribute defines the Attribute
type MachineGroupAttribute struct {
	ExternalName string `json:"externalName"`
	TopicName    string `json:"groupTopic"`
}

// MachineGroup defines the machine Group
type MachineGroup struct {
	Name          string   `json:"groupName"`
	Type          string   `json:"groupType"`
	MachineIDType string   `json:"machineIdentifyType"`
	MachineIDList []string `json:"machineList"`

	Attribute MachineGroupAttribute `json:"groupAttribute"`

	CreateTime     uint32
	LastModifyTime uint32

	project *LogProject
}

// Machine defines the Machine
type Machine struct {
	IP            string
	UniqueID      string `json:"machine-uniqueid"`
	UserdefinedID string `json:"userdefined-id"`
}

// MachineList defines the Machine List
type MachineList struct {
	Total    int
	Machines []*Machine
}

// ListMachines returns the machine list of this machine group.
func (m *MachineGroup) ListMachines() (ms []*Machine, total int, err error) {
	h := map[string]string{
		"x-sls-bodyrawsize": "0",
	}

	uri := fmt.Sprintf("/machinegroups/%v/machines", m.Name)
	r, err := request(m.project, "GET", uri, h, nil)
	if err != nil {
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	if r.StatusCode != http.StatusOK {
		errMsg := &errorMessage{}
		err = json.Unmarshal(buf, errMsg)
		if err != nil {
			err = fmt.Errorf("failed to remove config from machine group")
			dump, _ := httputil.DumpResponse(r, true)
			fmt.Println(dump)
			return
		}
		err = fmt.Errorf("%v:%v", errMsg.Code, errMsg.Message)
		return
	}

	body := &MachineList{}
	err = json.Unmarshal(buf, body)
	if err != nil {
		return
	}

	ms = body.Machines
	total = body.Total

	return
}

// GetAppliedConfigs returns applied configs of this machine group.
func (m *MachineGroup) GetAppliedConfigs() (confNames []string, err error) {
	confNames, err = m.project.GetAppliedConfigs(m.Name)
	return
}
