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

package config

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/roidelapluie/o11y-deploy/modules"
	"gopkg.in/yaml.v2"
)

var DefaultGlobal = Global{
	SDSyncTime:  model.Duration(10 * time.Second),
	DataDir:     "data",
	AnsibleTOFU: true,
	AnsibleUser: "ansible",
	EnableARA:   true,
	ARAListen:   "127.0.0.1:8089",
}

type Global struct {
	AnsibleSSHKeyPath         string         `yaml:"ansible_ssh_key_path"`
	AnsibleBecomePasswordFile string         `yaml:"ansible_become_password_file"`
	AnsibleUser               string         `yaml:"ansible_user"`
	SDSyncTime                model.Duration `yaml:"sd_sync_time"`
	DataDir                   string         `yaml:"data_directory"`
	AnsibleTOFU               bool           `yaml:"ansible_trust_on_firs_use"`
	EnableARA                 bool           `yaml:"enable_ara"`
	ARAListen                 string         `yaml:"ara_listen_address"`
}

var DefaultConfig = Config{}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (g *Global) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*g = DefaultGlobal
	// We want to set c to the defaults and then overwrite it with the input.
	// To make unmarshal fill the plain data struct rather than calling UnmarshalYAML
	// again, we have to hide it using a type indirection.
	type plain Global
	if err := unmarshal((*plain)(g)); err != nil {
		return err
	}
	return nil
}

func (g *Global) SetDirectory(directory string) {
	g.AnsibleSSHKeyPath = JoinDir(directory, g.AnsibleSSHKeyPath)
	g.DataDir = JoinDir(directory, g.DataDir)
}

type TargetGroup struct {
	Name    string   `yaml:"name"`
	Modules *Modules `yaml:"modules"`
	Targets *Targets `yaml:"targets"`
}

type Targets struct {
	ServiceDiscoveryConfigs discovery.Configs `yaml:"-"`
	RelabelConfigs          []*relabel.Config `yaml:"relabel_configs,omitempty"`
}

func (t *Targets) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := discovery.UnmarshalYAMLWithInlineConfigs(t, unmarshal); err != nil {
		return err
	}
	return nil
}

func (t *Targets) SetDirectory(directory string) {
	t.ServiceDiscoveryConfigs.SetDirectory(directory)
}

type Modules struct {
	ModulesConfigs modules.Configs `yaml:"-"`
}

func (t *TargetGroup) SetDirectory(directory string) {
	t.Targets.SetDirectory(directory)
}

func (m *Modules) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := modules.UnmarshalYAMLWithInlineConfigs(m, unmarshal); err != nil {
		return err
	}
	return nil
}

type Config struct {
	Global       Global        `yaml:"global"`
	TargetGroups []TargetGroup `yaml:"target_groups"`
}

func (c *Config) SetDirectory(directory string) {
	c.Global.SetDirectory(directory)
	for _, t := range c.TargetGroups {
		t.SetDirectory(directory)
	}
}

func LoadFile(filePath string) (*Config, error) {
	var config Config
	config = DefaultConfig

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = yaml.UnmarshalStrict(data, &config)
	if err != nil {
		return nil, err
	}

	config.SetDirectory(filepath.Dir(filePath))

	return &config, nil
}

func JoinDir(directory, relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(directory, relPath)
}
