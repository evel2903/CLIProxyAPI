package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
)

func TestAntigravityQuotaHTTPRouting(t *testing.T) {
	for _, model := range []string{"gemini-2.5-flash", "claude-sonnet-4-6"} {
		for _, stream := range []bool{false, true} {
			for _, known := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/stream=%t/known=%t", model, stream, known), func(t *testing.T) {
					var exhaustedCalls, healthyCalls atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.Header.Get("Authorization") == "Bearer exhausted" {
							exhaustedCalls.Add(1)
							w.WriteHeader(429)
							_, _ = w.Write([]byte(`{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Resource has been exhausted (e.g. check quota)."}}`))
							return
						}
						healthyCalls.Add(1)
						body := `{"response":{"candidates":[{"content":{"role":"model","parts":[{"text":"ok"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1,"totalTokenCount":2}}}`
						if stream {
							w.Header().Set("Content-Type", "text/event-stream")
							_, _ = fmt.Fprintf(w, "data: %s\n\n", body)
						} else {
							w.Header().Set("Content-Type", "application/json")
							_, _ = w.Write([]byte(body))
						}
					}))
					defer server.Close()
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					m := cliproxyauth.NewManager(nil, &cliproxyauth.FillFirstSelector{}, nil)
					m.RegisterExecutor(NewAntigravityExecutor(&config.Config{RequestRetry: 3}))
					for i, token := range []string{"exhausted", "healthy"} {
						id := fmt.Sprintf("quota-http-%d", i)
						registry.GetGlobalRegistry().RegisterClient(id, "antigravity", []*registry.ModelInfo{{ID: model}})
						t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(id) })
						_, err := m.Register(ctx, &cliproxyauth.Auth{ID: id, Provider: "antigravity", Status: cliproxyauth.StatusActive,
							Attributes: map[string]string{"base_url": server.URL},
							Metadata:   map[string]any{"access_token": token, "project_id": "test-project", "expired": time.Now().Add(time.Hour).Format(time.RFC3339)}})
						if err != nil {
							t.Fatal(err)
						}
					}
					if known {
						family := cliproxyauth.AntigravityQuotaFamily(model)
						body := []byte(fmt.Sprintf(`{"groups":[{"buckets":[{"bucketId":%q,"remainingFraction":0,"resetTime":%q}]}]}`, family+"-weekly", time.Now().Add(time.Hour).Format(time.RFC3339)))
						m.ApplyAntigravityQuotaSummary(ctx, "quota-http-0", body, time.Now())
					}
					req := cliproxyexecutor.Request{Model: model, Payload: []byte(`{"request":{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}}`)}
					opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatAntigravity}
					if stream {
						result, err := m.ExecuteStream(ctx, []string{"antigravity"}, req, opts)
						if err != nil {
							t.Fatal(err)
						}
						for chunk := range result.Chunks {
							if chunk.Err != nil {
								t.Fatal(chunk.Err)
							}
						}
					} else {
						if _, err := m.Execute(ctx, []string{"antigravity"}, req, opts); err != nil {
							t.Fatal(err)
						}
					}
					want := int32(1)
					if known {
						want = 0
					}
					if exhaustedCalls.Load() != want || healthyCalls.Load() != 1 {
						t.Fatalf("upstream calls exhausted=%d healthy=%d; want %d, 1", exhaustedCalls.Load(), healthyCalls.Load(), want)
					}
				})
			}
		}
	}
}

func TestAntigravityQuotaFetchUsesProjectAndCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if r.URL.Path != "/v1internal:retrieveUserQuotaSummary" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer quota-token" || body["project"] != "quota-project" {
			t.Error("incorrect quota request")
		}
		_, _ = w.Write([]byte(`{"groups":[]}`))
	}))
	defer server.Close()
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"base_url": server.URL}, Metadata: map[string]any{"access_token": "quota-token", "project_id": "quota-project", "expired": time.Now().Add(time.Hour).Format(time.RFC3339)}}
	if _, err := NewAntigravityExecutor(&config.Config{}).FetchQuotaSummary(context.Background(), auth); err != nil {
		t.Fatal(err)
	}
}
