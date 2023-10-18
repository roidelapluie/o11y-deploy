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

package portal

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/model/ctx"
	"github.com/roidelapluie/o11y-deploy/modules"
	"golang.org/x/crypto/bcrypt"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
)

var DefaultConfig = ModuleConfig{
	Enabled:      false,
	AuthpVersion: "1.0.3",
}

func init() {
	modules.RegisterConfig(&ModuleConfig{})
}

type ModuleConfig struct {
	Enabled      bool   `yaml:"enabled"`
	AuthpVersion string `yaml:"authp_version"`
	Users        []User `yaml:"users"`
}

type User struct {
	UUID           string `yaml:"uuid"`
	Username       string `yaml:"username"`
	BcryptPassword string `yaml:"bcrypt_password"`
	Email          string `yaml:"email"`
	Domain         string `yaml:"email_domain"`
	BcryptCost     int    `yaml:"bcrypt_cost"`
	Role           string `yaml:"role"`
}

func (m *ModuleConfig) Name() string {
	return "portal"
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
	users := m.cfg.Users

	var admin bool
	for i, user := range users {
		cost, err := bcrypt.Cost([]byte(user.BcryptPassword))
		if err != nil {
			return nil, fmt.Errorf("Error extracting cost for user %s: %v\n", user.Username, err)
		}
		if user.Role == "admin" {
			admin = true
		}
		users[i].BcryptCost = cost
		em := strings.Split(users[i].Email, "@")
		if len(em) > 1 {
			users[i].Domain = em[1]
		}
		users[i].UUID = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(user.Username)).String()
		if users[i].Role == "" {
			users[i].Role = "user"
		}
	}

	if !admin {
		dataDir := ctx.GetDatadir(c)
		if dataDir == "" {
			return nil, errors.New("Data directory not found")
		}
		pw, hash, err := getOrGetPassword(dataDir)
		if err != nil {
			return nil, err
		}
		if pw != "" {
			fmt.Printf("webadmin password is %s\n", pw)
		}
		users = append(users, User{
			Username:       "webadmin",
			BcryptPassword: hash,
			Email:          "admin@localhost",
			Role:           "admin",
		})
	}

	return &ansible.Playbook{
		Name: "Portal",
		Vars: map[string]interface{}{
			"authp_version":      m.cfg.AuthpVersion,
			"authp_users":        users,
			"o11y_proxy_entries": ctx.GetReverseProxyEntries(c),
		},
		Hosts:  "all",
		Become: true,
		Roles: []ansible.Role{
			{Name: "authp"},
		},
	}, nil
}

func (m *Module) GetTargets(labels []labels.Labels, group string) ([]labels.Labels, error) {
	return nil, nil
	// return modules.GetTargets(labels, "3000", group)
}

func (m *Module) HostVars(target labels.Labels, group string) (map[string]interface{}, error) {
	t := target.Copy()
	addr := t.Get(model.AddressLabel)
	if addr == "" {
		return nil, fmt.Errorf("__address__ label not found in label set")
	}
	host, _, err := net.SplitHostPort(string(addr))
	if err != nil {
		host = string(addr)
	}
	return map[string]interface{}{
		"o11y_portal_address": fmt.Sprintf("http://%s", host),
	}, nil
}
