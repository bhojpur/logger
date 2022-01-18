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

// InputDetail defines log detail
type InputDetail struct {
	LogType       string   `json:"logType"`
	LogPath       string   `json:"logPath"`
	FilePattern   string   `json:"filePattern"`
	LocalStorage  bool     `json:"localStorage"`
	TimeFormat    string   `json:"timeFormat"`
	LogBeginRegex string   `json:"logBeginRegex"`
	Regex         string   `json:"regex"`
	Keys          []string `json:"key"`
	FilterKeys    []string `json:"filterKey"`
	FilterRegex   []string `json:"filterRegex"`
	TopicFormat   string   `json:"topicFormat"`
}

// OutputDetail defines the output detail
type OutputDetail struct {
	Endpoint     string `json:"endpoint"`
	LogStoreName string `json:"logstoreName"`
}

// LogConfig defines Log Config
type LogConfig struct {
	Name         string       `json:"configName"`
	InputType    string       `json:"inputType"`
	InputDetail  InputDetail  `json:"inputDetail"`
	OutputType   string       `json:"outputType"`
	OutputDetail OutputDetail `json:"outputDetail"`

	CreateTime     uint32
	LastModifyTime uint32

	project *LogProject
}

// GetAppliedMachineGroup returns applied machine group of this config.
func (c *LogConfig) GetAppliedMachineGroup(confName string) (groupNames []string, err error) {
	groupNames, err = c.project.GetAppliedMachineGroups(c.Name)
	return
}
