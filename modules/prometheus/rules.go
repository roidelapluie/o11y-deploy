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

package prometheus

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
		Name: fmt.Sprintf("%v-prometheus", tg),
		Rules: []rulefmt.RuleNode{
			{
				Alert: node("PrometheusBadConfig"),
				Expr:  node("max_over_time(prometheus_config_last_reload_successful{job=\"prometheus\"}[5m]) == 0"),
				For:   model.Duration(10 * time.Minute),
				Annotations: map[string]string{
					"description": "Prometheus {{$labels.instance}} has failed to reload its configuration.",
					"summary":     "Failed Prometheus configuration reload.",
				},
				Labels: map[string]string{
					"severity": "critical",
				},
			},
			{
				Alert: node("PrometheusNotificationQueueRunningFull"),
				Expr: node(`
                    (
                        predict_linear(prometheus_notifications_queue_length{job="prometheus"}[5m], 60 * 30)
                    >
                        min_over_time(prometheus_notifications_queue_capacity{job="prometheus"}[5m])
                    )
                    `),
				For: model.Duration(15 * time.Minute),
				Annotations: map[string]string{
					"description": "Alert notification queue of Prometheus {{$labels.instance}} is running full.",
					"summary":     "Prometheus alert notification queue predicted to run full in less than 30m.",
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
