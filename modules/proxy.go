package modules

import "github.com/prometheus/prometheus/model/labels"

type ReverseProxiedModule interface {
	ReverseProxy([]labels.Labels, string) ([]ReverseProxyEntry, error)
}

type ReverseProxyEntry struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	Prefix string `yaml:"prefix"`
	Host   string `yaml:"host"`
}
