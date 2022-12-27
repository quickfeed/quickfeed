package manifest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestForm(t *testing.T) {
	tests := []struct {
		name       string
		domain     string
		status     int
		hasWebhook bool
	}{
		{
			name:       "no_webhook",
			domain:     "localhost",
			status:     http.StatusOK,
			hasWebhook: false,
		},
		{
			name:       "webhook",
			domain:     "example.com",
			status:     http.StatusOK,
			hasWebhook: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			if err := form(rr, tt.domain); err != nil {
				t.Fatalf("form() failed with error: %v", err)
			}
			if status := rr.Code; status != tt.status {
				t.Fatalf("form() returned wrong status code: got %v want %v", status, tt.status)
			}
			body := rr.Body.String()
			if tt.hasWebhook {
				if !strings.Contains(body, "default_events") {
					t.Errorf("form() returned body without default_events")
				}
				if !strings.Contains(body, "hook_attributes") {
					t.Errorf("form() returned body without hook_attributes")
				}
				if !strings.Contains(body, `"active": true`) {
					t.Errorf("form() returned body without active webhook")
				}
			} else {
				if strings.Contains(body, "default_events") {
					t.Errorf("form() returned body with default_events")
				}
				if strings.Contains(body, "hook_attributes") {
					t.Errorf("form() returned body with hook_attributes")
				}
				if strings.Contains(body, `"active": true`) {
					t.Errorf("form() returned body with active webhook")
				}
			}
			if t.Failed() {
				t.Log(body)
			}
		})
	}
}
