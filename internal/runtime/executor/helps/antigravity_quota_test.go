package helps

import "testing"

func TestAntigravityQuotaSwitchRequired(t *testing.T) {
	for _, test := range []struct {
		body string
		want bool
	}{
		{`{"error":{"status":"RESOURCE_EXHAUSTED"}}`, true},
		{`{"error":{"status":"RESOURCE_EXHAUSTED","details":[{"reason":"QUOTA_EXHAUSTED"}]}}`, true},
		{`{"error":{"status":"RESOURCE_EXHAUSTED","details":[{"reason":"RATE_LIMIT_EXCEEDED"}]}}`, false},
		{`{"error":{"status":"UNAVAILABLE"}}`, false},
		{`not json`, false},
	} {
		if got := AntigravityQuotaSwitchRequired([]byte(test.body)); got != test.want {
			t.Fatalf("%s: got %v", test.body, got)
		}
	}
}
