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
	"regexp"
	"strings"

	"github.com/roidelapluie/o11y-deploy/model/ansible"
	"github.com/roidelapluie/o11y-deploy/model/ctx"
	"github.com/roidelapluie/o11y-deploy/model/dashboard"
	"github.com/roidelapluie/o11y-deploy/modules"
	"github.com/roidelapluie/o11y-deploy/util"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

var DefaultConfig = ModuleConfig{
	AdminPassword:     "changeme",
	Enabled:           false,
	GrafanaVersion:    "10.2.1",
	DashboardsDir:     "/usr/share/o11y-dashboards",
	GrafanaAddress:    "127.0.0.1",
	GrafanaPort:       3000,
	AutoAssignOrgRole: "Viewer",
}

func init() {
	modules.RegisterConfig(&ModuleConfig{})
}

type ModuleConfig struct {
	AdminPassword     string `yaml:"admin_password"`
	Enabled           bool   `yaml:"enabled"`
	GrafanaVersion    string `yaml:"grafana_version"`
	DashboardsDir     string `yaml:"o11y_dashboards_dir"`
	GrafanaAddress    string `yaml:"grafana_address"`
	GrafanaPort       int64  `yaml:"grafana_port"`
	AutoAssignOrgRole string `yaml:"users_role"`
}

type GrafanaServerConfig struct {
	GrafanaServer    string `yaml:"grafana_server"`
	Protocol         string `yaml:"protocol"`
	EnforceDomain    bool   `yaml:"enforce_domain"`
	Socket           string `yaml:"socket"`
	CertKey          string `yaml:"cert_key"`
	CertFile         string `yaml:"cert_file"`
	EnableGzip       bool   `yaml:"enable_gzip"`
	StaticRootPath   string `yaml:"static_root_path"`
	RouterLogging    bool   `yaml:"router_logging"`
	ServeFromSubPath bool   `yaml:"serve_from_sub_path"`
}

type GrafanaUsers struct {
	AllowSignUp       bool   `yaml:"allow_sign_up"`
	AutoAssignOrgRole string `yaml:"auto_assign_org_role"`
	DefaultTheme      string `yaml:"default_theme"`
	// AllowOrgCreate         bool   `yaml:"allow_org_create"`
	// AutoAssignOrg          bool   `yaml:"auto_assign_org"`
	// LoginHint              string `yaml:"login_hint"`
	// ExternalManageLinkURL  string `yaml:"external_manage_link_url"`
	// ExternalManageLinkName string `yaml:"external_manage_link_name"`
	// ExternalManageInfo     string `yaml:"external_manage_info"`
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

		for i, _ := range dashboar.Templating.List {
			dashboar.Templating.List[i].Datasource = dashboard.TemplatingDataSource{
				Type: "prometheus",
				UID:  "${prometheus_ds}",
			}
		}

		if len(dashboar.Templating.List) > 0 {
			det := deepCopyTemplatingDetail(dashboar.Templating.List[0])
			expr := det.Query.ObjectValue.Query
			var err error
			det.Name = "group_name"
			det.Query.ObjectValue.Query, err = recodeQuery(expr, "", "group_name")
			if err != nil {
				return nil, err
			}

			dashboar.Templating.List = append([]dashboard.TemplatingDetail{
				det,
			}, dashboar.Templating.List...)
		}

		dashboar.Templating.List = append([]dashboard.TemplatingDetail{
			{
				Hide:       1,
				IncludeAll: false,
				Label:      "",
				Multi:      false,
				Name:       "prometheus_ds",
				Options:    []interface{}{},
				Query: dashboard.QueryValue{
					StringValue: "prometheus",
					IsString:    true,
				},
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
				xp, err := addGroupNameSelector(nt.Expr)
				if err != nil {
					return nil, err
				}
				nt.Expr = xp
				np.Targets[j] = nt
			}
			dashboar.Panels[i] = np
		}

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
			"grafana_address":        m.cfg.GrafanaAddress,
			"grafana_port":           m.cfg.GrafanaPort,
			"grafana_datasources":    grafanaDS,
			"grafana_dashboards_dir": directoryPath,
			"grafana_metrics": map[string]interface{}{
				"enabled": true,
			},
			"grafana_auth": map[string]interface{}{
				"disable_login_form":   true,
				"oauth_auto_login":     false,
				"disable_signout_menu": false,
				"signout_redirect_url": "/auth/logout",
				"proxy": map[string]interface{}{
					"enabled":         true,
					"header_name":     "X-Token-Subject",
					"header_property": "username",
					"auto_sign_up":    true,
				},
			},
			"grafana_server": GrafanaServerConfig{
				GrafanaServer:    "localhost",
				Protocol:         "http",
				EnforceDomain:    false,
				Socket:           "",
				CertKey:          "",
				CertFile:         "",
				EnableGzip:       false,
				StaticRootPath:   "public",
				RouterLogging:    false,
				ServeFromSubPath: true,
			},
			"grafana_users": GrafanaUsers{
				AllowSignUp:       false,
				AutoAssignOrgRole: m.cfg.AutoAssignOrgRole,
				DefaultTheme:      "dark",
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
	addr, err := modules.GetReverseProxyAddress(target, m.cfg.Name(), "/grafana", group)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"grafana_url": addr,
	}, nil
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

func addGroupNameSelector(query string) (string, error) {
	expr, err := parser.ParseExpr(encodeGrafanaVar(query))
	if err != nil {
		return query, err
	}
	name := "group_name"
	value := "$group_name"
	matchType := labels.MatchRegexp

	parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
		if n, ok := node.(*parser.VectorSelector); ok {
			var found bool
			for i, l := range n.LabelMatchers {
				if l.Name == name {
					n.LabelMatchers[i].Type = matchType
					n.LabelMatchers[i].Value = value
					found = true
				}
			}
			if !found {
				n.LabelMatchers = append(n.LabelMatchers, &labels.Matcher{
					Type:  matchType,
					Name:  name,
					Value: value,
				})
			}
		}
		return nil
	})
	if err != nil {
		return query, err
	}
	return decodeGrafanaVar(expr.Pretty(0)), nil
}

// extractQuery takes a Grafana-style label_values query and extracts the
// metric and its selectors, then constructs the desired query format.
func extractQuery(query string) (string, error) {
	// Regular expression to match the query format
	// e.g., label_values(node_exporter_build_info{fo="bar",bar="foo",ss="xx"},instance)
	re := regexp.MustCompile(`label_values\((?P<metric>[a-zA-Z_][a-zA-Z0-9_]*)({(?P<labels>.*)})?,((?P<label>[a-zA-Z_][a-zA-Z0-9_]*))\)`)
	matches := re.FindStringSubmatch(query)

	if len(matches) == 0 {
		return "", fmt.Errorf("invalid query format")
	}

	metric := matches[re.SubexpIndex("metric")]
	labels := matches[re.SubexpIndex("labels")]

	// Construct the desired query format
	if labels != "" {
		return fmt.Sprintf("%s{%s}", metric, labels), nil
	}
	return metric, nil
}

func recodeQuery(query, newMetric, newLabel string) (string, error) {
	// Regular expression to match the query format
	// e.g., label_values(node_exporter_build_info{fo="bar",bar="foo",ss="xx"},instance)
	re := regexp.MustCompile(`label_values\((?P<metric>[a-zA-Z_][a-zA-Z0-9_]*)({(?P<labels>.*)})?,((?P<label>[a-zA-Z_][a-zA-Z0-9_]*))\)`)
	matches := re.FindStringSubmatch(query)

	if len(matches) == 0 {
		return "", fmt.Errorf("invalid query format")
	}

	label := matches[re.SubexpIndex("label")]
	metric := matches[re.SubexpIndex("metric")]
	labels := matches[re.SubexpIndex("labels")]
	if labels != "" {
		metric = fmt.Sprintf("%s{%s}", metric, labels)
	}
	if newLabel != "" {
		return fmt.Sprintf("label_values(%s,%s)", metric, newLabel), nil
	}
	if newMetric != "" {
		return fmt.Sprintf("label_values(%s,%s)", newMetric, label), nil
	}

	// Construct the desired query format
	return fmt.Sprintf("label_values(%s,%s)", metric, label), nil
}

func deepCopyTemplatingDetail(src dashboard.TemplatingDetail) dashboard.TemplatingDetail {
	// Start with a shallow copy
	cpy := src

	// Directly initialize all fields for newQueryValue
	newQueryValue := dashboard.QueryValue{
		StringValue: src.Query.StringValue,
		ObjectValue: &dashboard.QueryObject{
			Query: src.Query.ObjectValue.Query,
			Refid: src.Query.ObjectValue.Refid,
		},
		IsString: src.Query.IsString,
	}

	cpy.Query = newQueryValue

	return cpy
}

func (m *Module) ReverseProxy(targets []labels.Labels, group string) ([]modules.ReverseProxyEntry, error) {
	rp, err := modules.GetReverseProxy(targets, fmt.Sprintf("%d", m.cfg.GrafanaPort), m.cfg.Name(), "/grafana", group)
	if err != nil {
		return rp, fmt.Errorf("could not get reverse proxy entries: %v", err)
	}

	// If we're only listening on localhost, replace the host part of the entry.
	if m.cfg.GrafanaAddress == "127.0.0.1" {
		rp, err = util.ReplaceHost(rp, "127.0.0.1")
		if err != nil {
			return rp, fmt.Errorf("could not replace host address with localhost: %v", err)
		}
	}

	return rp, nil
}
