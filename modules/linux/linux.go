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
	"time"

	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/modules"
	"gopkg.in/yaml.v3"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/rulefmt"
)

var DefaultConfig = ModuleConfig{
	EnableExporter:      true,
	NodeExporterVersion: "1.5.0",
}

func init() {
	modules.RegisterConfig(&ModuleConfig{})
}

type ModuleConfig struct {
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

func node(value string) yaml.Node {
	return yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
	}
}

// GetRules returns recording and alerting rules for this module.
func (m *Module) GetRules() rulefmt.RuleGroup {
	return rulefmt.RuleGroup{
		Name: "linux",
		Rules: []rulefmt.RuleNode{
			{
				Alert: node("HostOutOfMemory"),
				Expr:  node("(node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100 < 10) * on(instance) group_left (nodename) node_uname_info{nodename=~\".+\"}"),
				For:   model.Duration(2 * time.Minute),
				Annotations: map[string]string{
					"summary":     "Host out of memory (instance {{ $labels.instance }})",
					"description": "Node memory is filling up (< 10% left)\n  VALUE = {{ $value }}\n  LABELS = {{ $labels }}",
				},
				Labels: map[string]string{
					"severity": "warning",
				},
			},
		},
	}
}

func (m *Module) HostVars() (map[string]string, error) {
	return nil, nil
}
