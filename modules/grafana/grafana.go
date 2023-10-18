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

package grafana

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/model/ctx"
	"github.com/roidelapluie/o11y-deploy/modules"

	"github.com/prometheus/prometheus/model/labels"
)

var DefaultConfig = ModuleConfig{
	AdminPassword:  "changeme",
	Enabled:        false,
	GrafanaVersion: "latest",
}

func init() {
	modules.RegisterConfig(&ModuleConfig{})
}

type ModuleConfig struct {
	AdminPassword  string `yaml:"admin_password"`
	Enabled        bool   `yaml:"enabled"`
	GrafanaVersion string `yaml:"grafana_version"`
}

func (m *ModuleConfig) Name() string {
	return "grafana"
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

func (m *Module) Playbook(c context.Context) (*ansible.Playbook, error) {
	dashboardsDir, err := os.MkdirTemp("", "grafana-dashboards")
	if err != nil {
		return nil, fmt.Errorf("could not create dashboard directory: %v", err)
	}

	dashboardFiles := ctx.GetDashboardFiles(c)
	for name, content := range dashboardFiles {
		err = os.WriteFile(filepath.Join(dashboardsDir, name), content, 0666)
	}

	return &ansible.Playbook{
		Name: "Grafana",
		Vars: map[string]interface{}{
			"grafana_version": m.cfg.GrafanaVersion,
			"grafana_security": map[string]string{
				"admin_user":     "admin",
				"admin_password": m.cfg.AdminPassword,
			},
			"grafana_datasources": []map[string]interface{}{
				{
					"name":       "prometheus",
					"type":       "prometheus",
					"access":     "proxy",
					"url":        fmt.Sprintf("http://%s:9090/", strings.Split(ctx.GetPromServers(c)[0], ":")[0]),
					"basic_auth": false,
				},
			},
			"grafana_dashboards":     []string{}, //  ctx.GetDashboards(c),
			"grafana_dashboards_dir": dashboardsDir,
			"grafana_metrics": map[string]interface{}{
				"enabled": true,
			},
		},
		Hosts:  "all",
		Become: true,
		Roles: []ansible.Role{
			{Name: "grafana"},
		},
	}, nil
}

func (m *Module) GetTargets(labels []labels.Labels, group string) ([]labels.Labels, error) {
	return modules.GetTargets(labels, "3000", group)
}

func (m *Module) HostVars(target labels.Labels, group string) (map[string]interface{}, error) {
	return nil, nil
}
