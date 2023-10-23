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
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/model/ctx"
	"github.com/roidelapluie/o11y-deploy/model/dashboard"
	"github.com/roidelapluie/o11y-deploy/modules"

	"github.com/prometheus/prometheus/model/labels"
)

var DefaultConfig = ModuleConfig{
	AdminPassword:  "changeme",
	Enabled:        false,
	GrafanaVersion: "latest",
	DashboardsDir:  "/usr/share/o11y-dashboards",
}

func init() {
	modules.RegisterConfig(&ModuleConfig{})
}

type ModuleConfig struct {
	AdminPassword  string `yaml:"admin_password"`
	Enabled        bool   `yaml:"enabled"`
	GrafanaVersion string `yaml:"grafana_version"`
	DashboardsDir  string `yaml:"o11y_dashboards_dir"`
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

	ctxDashboards := ctx.GetDashboards(c)
	dir := ctx.GetDatadir(c)
	directoryPath := dir + "/dashboards"
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		err := os.Mkdir(directoryPath, 0755)
		if err != nil {
			return nil, err
		}
	}

	expectedFiles := make(map[string]bool)

	for _, d := range ctxDashboards {
		var dashboar dashboard.Dashboard
		err := json.Unmarshal(d, &dashboar)
		if err != nil {
			return nil, err
		}

		dashboar.Templating.List = append([]dashboard.TemplatingDetail{
			{
				Hide:        1,
				IncludeAll:  false,
				Label:       "",
				Multi:       false,
				Name:        "prometheus_ds",
				Options:     []interface{}{},
				Query:       "prometheus",
				QueryValue:  "",
				Refresh:     1,
				Regex:       "",
				SkipUrlSync: false,
				Type:        "datasource",
			},
		}, dashboar.Templating.List...)

		for i, p := range dashboar.Panels {
			np := p
			np.Datasource.UID = "${prometheus_ds}"
			for j, t := range np.Targets {
				nt := t
				nt.Datasource.UID = "${prometheus_ds}"
				np.Targets[j] = nt
			}
			dashboar.Panels[i] = np
		}

		//escapeStructStrings(reflect.ValueOf(&dashboar).Elem())

		hasher := sha1.New()
		hasher.Write([]byte(dashboar.Title))
		filename := hex.EncodeToString(hasher.Sum(nil)) + ".json"
		expectedFiles[filename] = true

		filePath := directoryPath + "/" + filename
		fileContent, err := json.Marshal(dashboar)
		if err != nil {
			return nil, fmt.Errorf("error marshalling dashboard: %v", err)
		}
		err = ioutil.WriteFile(filePath, fileContent, 0644)
		if err != nil {
			return nil, fmt.Errorf("error writing to file: %v", err)
		}
	}
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		baseName := filepath.Base(file.Name())
		if _, exists := expectedFiles[baseName]; !exists {
			os.Remove(directoryPath + "/" + file.Name())
		}
	}

	grafanaDS := make([]map[string]interface{}, 0)
	for _, s := range ctx.GetPromServers(c) {
		grafanaDS = append(grafanaDS, map[string]interface{}{
			"name":       s.Name,
			"type":       "prometheus",
			"access":     "proxy",
			"url":        s.URL,
			"basic_auth": false,
		})
	}

	return &ansible.Playbook{
		Name: "Grafana",
		Vars: map[string]interface{}{
			"grafana_version":             m.cfg.GrafanaVersion,
			"grafana_provisioning_synced": true,
			"grafana_security": map[string]string{
				"admin_user":     "admin",
				"admin_password": m.cfg.AdminPassword,
			},
			"grafana_datasources":    grafanaDS,
			"grafana_dashboards_dir": directoryPath,
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

func replaceBraces(data []byte) []byte {
	data = bytes.Replace(data, []byte("{{"), []byte(`{{ "{{" }}`), -1)
	data = bytes.Replace(data, []byte("}}"), []byte(`{{ "}}" }}`), -1)
	return data
}

func escapeTemplates(input string) string {
	escapedStart := "{% raw %}{{"
	escapedEnd := "}}{% endraw %}"
	escapedContent := strings.ReplaceAll(input, "{{", escapedStart)
	escapedContent = strings.ReplaceAll(escapedContent, "}}", escapedEnd)
	return escapedContent
}

func escapeStructStrings(v reflect.Value) {
	t := v.Type()

	// If it's a pointer, we need to dereference it.
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		switch field.Kind() {
		case reflect.String:
			if field.CanSet() {
				escapedString := escapeTemplates(field.String())
				field.SetString(escapedString)
			}
		case reflect.Struct:
			escapeStructStrings(field)
		case reflect.Slice:
			elemType := t.Field(i).Type.Elem()
			if elemType.Kind() == reflect.Struct || (elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct) {
				for j := 0; j < field.Len(); j++ {
					escapeStructStrings(field.Index(j))
				}
			}
		}
	}
}
