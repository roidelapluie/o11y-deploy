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

package alertmanager

import (
	"context"
	"fmt"
	"net"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/roidelapluie/o11y-deploy/model/amserver"
	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/modules"
	"github.com/roidelapluie/o11y-deploy/util"
)

var DefaultConfig = ModuleConfig{
	Enabled:       false,
	ListenAddress: "127.0.0.1",
	ListenPort:    "9093",
	Receivers: []string{
		"default@change.me",
	},
	SmtpFrom:      "default@change.me",
	SmtpSmarthost: "smtp.gmail.com:587",
}

func init() {
	modules.RegisterConfig(&ModuleConfig{})
}

type ModuleConfig struct {
	Enabled       bool     `yaml:"enabled"`
	ListenAddress string   `yaml:"listen_address"`
	ListenPort    string   `yaml:"listen_port"`
	Receivers     []string `yaml:"receivers"`
	SmtpFrom      string   `yaml:"smtp_from"`
	SmtpSmarthost string   `yaml:"smtp_smarthost"`
}

func (m *ModuleConfig) Name() string {
	return "alertmanager"
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

func mapEmailsToConfig(emails []string) []interface{} {
	result := []interface{}{}
	for _, email := range emails {
		result = append(result, map[string]string{
			"to": email,
		})
	}
	return result
}

func (m *Module) Playbook(c context.Context) (*ansible.Playbook, error) {

	return &ansible.Playbook{
		Name: "Alertmanager",
		Vars: map[string]interface{}{
			"alertmanager_receivers": []map[string]interface{}{
				{
					"name":          "email",
					"email_configs": mapEmailsToConfig(m.cfg.Receivers),
				},
			},
			"alertmanager_route": map[string]interface{}{
				"group_by":        []string{"alertname"},
				"group_wait":      "30s",
				"group_interval":  "5m",
				"repeat_interval": "3h",
				"receiver":        "email",
			},
			"alertmanager_smtp": map[string]interface{}{
				"from":      m.cfg.SmtpFrom,
				"smarthost": m.cfg.SmtpSmarthost,
			},
			"alertmanager_web_external_url":   "{{o11y_alertmanager_external_address}}",
			"alertmanager_web_listen_address": net.JoinHostPort(m.cfg.ListenAddress, m.cfg.ListenPort),
		},
		Hosts:  "all",
		Become: true,
		Roles: []ansible.Role{
			{
				Name: "alertmanager",
			},
		},
	}, nil
}

func (m *Module) HostVars(target labels.Labels, group string) (map[string]interface{}, error) {
	addr, err := modules.GetReverseProxyAddress(target, m.cfg.Name(), "/alertmanager", group)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"o11y_alertmanager_external_address": addr,
	}, nil
}

func (m *Module) GetTargets(targets []labels.Labels, group string) ([]labels.Labels, error) {
	return modules.GetTargets(targets, m.cfg.ListenPort, group)
}

func (m *Module) ReverseProxy(targets []labels.Labels, group string) ([]modules.ReverseProxyEntry, error) {
	rp, err := modules.GetReverseProxy(targets, m.cfg.ListenPort, m.cfg.Name(), "/alertmanager", group)
	if err != nil {
		return rp, fmt.Errorf("could not get reverse proxy entries: %v", err)
	}

	// If we're only listening on localhost, replace the host part of the entry.
	if m.cfg.ListenAddress == "127.0.0.1" {
		rp, err = util.ReplaceHost(rp, "127.0.0.1")
		if err != nil {
			return rp, fmt.Errorf("could not replace host address with localhost: %v", err)
		}
	}

	return rp, nil
}

func (m *Module) GetAlertmanagerServers(targets []labels.Labels, group string) ([]amserver.AlertmanagerServer, error) {
	rp, err := modules.GetReverseProxy(targets, m.cfg.ListenPort, m.cfg.Name(), "/alertmanager", group)
	if err != nil {
		return nil, err
	}
	amservers := []amserver.AlertmanagerServer{}
	for _, r := range rp {
		amservers = append(amservers, amserver.AlertmanagerServer{
			Name: r.Name,
			URL:  r.URL + r.Prefix,
		})
	}
	return amservers, nil
}
