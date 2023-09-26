// Copyright 2023 The O11y Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ansible

type Inventory struct {
	Groups map[string]Group `yaml:",inline"`
}

type Group struct {
	Hosts     map[string]Host        `yaml:"hosts,omitempty"`
	Variables map[string]interface{} `yaml:"vars,omitempty"`
}

type Host struct {
	Variables map[string]interface{} `yaml:"host_var,omitempty"`
}

type Hosts struct {
}

type Playbook struct {
	Name       string                 `yaml:"name"`
	Hosts      string                 `yaml:"hosts,omitempty"`
	RemoteUser string                 `yaml:"remote_user,omitempty"`
	Vars       map[string]interface{} `yaml:"vars,omitempty"`
	Tasks      []Task                 `yaml:"tasks"`
	Roles      []Role                 `yaml:"roles,omitempty"`
	Become     bool                   `yaml:"become"`
}

type Task struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:",inline"`
}

type Role struct {
	Name  string `yaml:"name"`
	Tasks []Task `yaml:"tasks,omitempty"`
}
