package helps

import (
	"encoding/json"
	"strings"
)

// AntigravityQuotaSwitchRequired separates exhausted resource budgets from
// explicit short rate limits. Bare RESOURCE_EXHAUSTED is ambiguous about the
// reset time, but retrying that same credential delays otherwise healthy ones.
// The manager applies a bounded cooldown while quota polling discovers reset.
func AntigravityQuotaSwitchRequired(body []byte) bool {
	var response struct {
		Error struct {
			Status  string `json:"status"`
			Details []struct {
				Reason string `json:"reason"`
			} `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &response) != nil || !strings.EqualFold(response.Error.Status, "RESOURCE_EXHAUSTED") {
		return false
	}
	for _, detail := range response.Error.Details {
		if strings.EqualFold(detail.Reason, "QUOTA_EXHAUSTED") {
			return true
		}
	}
	for _, detail := range response.Error.Details {
		if strings.EqualFold(detail.Reason, "RATE_LIMIT_EXCEEDED") {
			return false
		}
	}
	return true
}
