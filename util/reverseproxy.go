package util

import (
	"fmt"
	"net"
	"net/url"

	"github.com/roidelapluie/o11y-deploy/modules"
)

// ReplaceHost replaces the host part in a set of ReverseProxyEntries
func ReplaceHost(entries []modules.ReverseProxyEntry, host string) ([]modules.ReverseProxyEntry, error) {
	for i := range entries {
		u, err := url.Parse(entries[i].URL)
		if err != nil {
			return entries, fmt.Errorf("could not parse url: %v", err)
		}

		entries[i].Host = host
		entries[i].URL = fmt.Sprintf("%s://%s", u.Scheme, net.JoinHostPort(host, u.Port()))
	}

	return entries, nil
}
