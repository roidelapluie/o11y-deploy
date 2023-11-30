package dashboard

import (
	_ "embed"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"reflect"
	"github.com/google/go-cmp/cmp"
)

//go:embed Linux.json
var data []byte

func TestUnmarshalAndCompareDashboard(t *testing.T) {
	var dashboard Dashboard
	err := json.Unmarshal(data, &dashboard)
	if err != nil {
		t.Fatalf("Failed to unmarshal to Dashboard: %v", err)
	}

	// Unmarshal to an empty interface
	var genericData interface{}
	err = json.Unmarshal(data, &genericData)
	if err != nil {
		t.Fatalf("Failed to unmarshal to interface{}: %v", err)
	}

	// Pretty marshal both to temporary files
	dashboardData, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		t.Fatalf("Failed to pretty marshal dashboard data: %v", err)
	}

	genericDataData, err := json.MarshalIndent(genericData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to pretty marshal generic data: %v", err)
	}

	// Write to temp files
	dashboardFile, err := ioutil.TempFile("", "dashboard")
	if err != nil {
		t.Fatalf("Failed to create temp file for dashboard: %v", err)
	}
	defer os.Remove(dashboardFile.Name())

	genericDataFile, err := ioutil.TempFile("", "genericData")
	if err != nil {
		t.Fatalf("Failed to create temp file for generic data: %v", err)
	}
	defer os.Remove(genericDataFile.Name())

	_, err = dashboardFile.Write(dashboardData)
	if err != nil {
		t.Fatalf("Failed to write to dashboard temp file: %v", err)
	}

	_, err = genericDataFile.Write(genericDataData)
	if err != nil {
		t.Fatalf("Failed to write to generic data temp file: %v", err)
	}

	// Compare the two files
	if !reflect.DeepEqual(dashboardData, genericDataData) {
		diff := cmp.Diff(string(dashboardData), string(genericDataData))
		t.Errorf("Unmarshalled data differs between Dashboard struct and interface{}.\nDiff:\n%s", diff)
	}
}

