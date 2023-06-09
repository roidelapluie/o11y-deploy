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

package linux

import (
	"context"

	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/modules"

	"github.com/prometheus/prometheus/model/labels"
)

var DefaultConfig = ModuleConfig{
	Enabled:             false,
	EnableExporter:      true,
	NodeExporterVersion: "1.5.0",
}

func init() {
	modules.RegisterConfig(&ModuleConfig{})
}

type ModuleConfig struct {
	Enabled             bool   `yaml:"enabled"`
	EnableExporter      bool   `yaml:"enable_exporter"`
	NodeExporterVersion string `yaml:"node_exporter_version"`
}

func (m *ModuleConfig) Name() string {
	return "linux"
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (m *ModuleConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*m = DefaultConfig
	// We want to set m to the defaults and then overwrite it with the input.
	// To make unmarshal fill the plain data struct rather than calling UnmarshalYAML
	// again, we have to hide it using a type indirection.
	type plain ModuleConfig
	if err := unmarshal((*plain)(m)); err != nil {
		return err
	}
	return nil
}

func (m *ModuleConfig) NewModule(modules.ModuleOptions) (modules.Module, error) {
	return &Module{
		cfg: m,
	}, nil
}

// IsEnabled returns a boolean indicating if the Module is enabled.
func (m *ModuleConfig) IsEnabled() bool {
	return m.Enabled
}

type Module struct {
	cfg *ModuleConfig
}

func (m *Module) Playbook(ctx context.Context) (*ansible.Playbook, error) {
	if !m.cfg.EnableExporter {
		return nil, nil
	}
	return &ansible.Playbook{
		Name: "Linux",
		Vars: map[string]interface{}{
			"node_exporter_version": m.cfg.NodeExporterVersion,
		},
		Hosts:  "all",
		Become: true,
		Roles: []ansible.Role{
			ansible.Role{
				Name: "node_exporter",
			},
		},
	}, nil
}

func (m *Module) GetTargets(labels []labels.Labels) ([]labels.Labels, error) {
	return modules.GetTargets(labels)
}

func (m *Module) HostVars() (map[string]string, error) {
	return nil, nil
}
