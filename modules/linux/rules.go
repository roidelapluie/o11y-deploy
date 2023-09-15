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
	"fmt"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
)

// GetRules returns recording and alerting rules for this module.
func (m *Module) GetRules(tg string) rulefmt.RuleGroup {
	return rulefmt.RuleGroup{
		Name: fmt.Sprintf("%v-linux", tg),
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

func node(value string) yaml.Node {
	return yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
	}
}
