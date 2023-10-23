package dashboard

import (
	_ "embed"
	"encoding/json"
	"testing"
)

//go:embed Linux.json
var data []byte

func TestUnmarshalDashboard(t *testing.T) {

	var dashboard Dashboard
	err := json.Unmarshal(data, &dashboard)
	if err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	// Sample assertions
	if dashboard.ID != 1 {
		t.Errorf("Expected ID to be 1, but got %d", dashboard.ID)
	}

	if dashboard.Title != "Linux" {
		t.Errorf("Expected Title to be 'Linux', but got '%s'", dashboard.Title)
	}

}
