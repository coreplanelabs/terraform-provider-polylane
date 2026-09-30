package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
)

func TestGCPActivationRetry(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		statuses  []int
		wantCalls int
		wantError bool
	}{
		{"transient eventually registers", []int{503, 429, 200}, 3, false},
		{"wrong identity is terminal", []int{400}, 1, true},
		{"unauthorized is terminal", []int{403}, 1, true},
		{"conflict is terminal", []int{409}, 1, true},
		{"transient budget exhausted", []int{503}, 6, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			waits := 0
			c, err := client.New("test", "https://example.invalid/v1", "test", &http.Client{Transport: gcpProviderTransport(func(r *http.Request) (*http.Response, error) {
				status := tt.statuses[min(calls, len(tt.statuses)-1)]
				calls++
				b, _ := io.ReadAll(r.Body)
				if !strings.Contains(string(b), `"requestId":"request"`) {
					t.Error("retry changed request identity")
				}
				body := `{"success":false,"message":"unavailable"}`
				if status == 200 {
					body = `{"success":true,"result":{"requestId":"request","cloudAccountId":"account","provisioningStatus":"registered"}}`
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			_, err = activateGCPWithRetry(context.Background(), c, client.ActivateGCPConnectionInput{WorkspaceID: "workspace", RequestID: "request", ProjectNumber: "123456789"}, func(context.Context, time.Duration) error { waits++; return nil })
			if (err != nil) != tt.wantError || calls != tt.wantCalls || waits != calls-1 {
				t.Fatalf("calls=%d waits=%d error=%v", calls, waits, err)
			}
		})
	}
}
