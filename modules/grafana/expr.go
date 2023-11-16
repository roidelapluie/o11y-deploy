package grafana

import "strings"

// Define the mapping of Grafana variables to specific hours/times
var grafanaMapping = map[string]string{
	"$__rate_interval": "999m",
	"$__range":         "998m",
}

// Encode function to convert Grafana variables to specific hours/times
func encodeGrafanaVar(query string) string {
	for k, v := range grafanaMapping {
		query = strings.ReplaceAll(query, k, v)
	}
	return query
}

// Decode function to revert specific hours/times back to Grafana variables
func decodeGrafanaVar(encodedQuery string) string {
	for k, v := range grafanaMapping {
		encodedQuery = strings.ReplaceAll(encodedQuery, v, k)
	}
	return encodedQuery
}
