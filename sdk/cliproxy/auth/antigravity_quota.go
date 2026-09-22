package auth

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

const antigravityQuotaPrefix = "@quota/antigravity/"

// QuotaSummaryFetcher provides authoritative quota windows without generating tokens.
type QuotaSummaryFetcher interface {
	FetchQuotaSummary(context.Context, *Auth) ([]byte, error)
}

// AntigravityQuotaFamily maps upstream model IDs, never UI display labels.
func AntigravityQuotaFamily(model string) string {
	model = strings.ToLower(canonicalModelKey(model))
	switch {
	case strings.HasPrefix(model, "gemini-"):
		return "gemini"
	case strings.HasPrefix(model, "claude-"), strings.HasPrefix(model, "gpt-"):
		return "3p"
	default:
		return ""
	}
}

// ParseAntigravityQuotaSummary returns only known families with valid observations.
// A zero time means the observed windows are available, not an unknown response.
// Both weekly and short windows constrain a family: the latest exhausted reset wins.
func ParseAntigravityQuotaSummary(body []byte, now time.Time) map[string]time.Time {
	var response struct {
		Groups []struct {
			Buckets []struct {
				ID        string   `json:"bucketId"`
				Remaining *float64 `json:"remainingFraction"`
				Reset     string   `json:"resetTime"`
				Disabled  bool     `json:"disabled"`
			} `json:"buckets"`
		} `json:"groups"`
	}
	if json.Unmarshal(body, &response) != nil {
		return nil
	}
	result := make(map[string]time.Time)
	invalid := make(map[string]bool)
	seen := make(map[string]bool)
	for _, group := range response.Groups {
		for _, bucket := range group.Buckets {
			family := ""
			switch bucket.ID {
			case "gemini-weekly", "gemini-5h":
				family = "gemini"
			case "3p-weekly", "3p-5h":
				family = "3p"
			}
			if family == "" {
				continue
			}
			seen[bucket.ID] = true
			if bucket.Disabled {
				continue
			}
			if bucket.Remaining == nil || *bucket.Remaining < 0 || *bucket.Remaining > 1 {
				invalid[family] = true
				continue
			}
			if _, exists := result[family]; !exists {
				result[family] = time.Time{}
			}
			if *bucket.Remaining > 0 {
				continue
			}
			reset, err := time.Parse(time.RFC3339Nano, bucket.Reset)
			if err != nil || !reset.After(now) {
				invalid[family] = true
				continue
			}
			if reset.After(result[family]) {
				result[family] = reset
			}
		}
	}
	for family, reset := range result {
		// A partial response cannot clear a previously exhausted sibling window.
		// Preserve valid exhaustion even when a sibling is missing or malformed.
		if reset.IsZero() && (invalid[family] || !seen[family+"-weekly"] || !seen[family+"-5h"]) {
			delete(result, family)
		}
	}
	return result
}

// ApplyAntigravityQuotaSummary updates the same scheduler snapshots used by all
// selectors. Group state is separate from individual request results, so an
// in-flight success cannot erase a newly observed weekly quota limit.
func (m *Manager) ApplyAntigravityQuotaSummary(ctx context.Context, authID string, body []byte, observedAt time.Time) {
	observed := ParseAntigravityQuotaSummary(body, observedAt)
	if m == nil || len(observed) == 0 {
		return
	}
	m.mu.Lock()
	auth := m.auths[authID]
	if auth == nil || auth.Provider != "antigravity" || auth.Disabled || m.cooldownDisabledForAuth(auth) {
		m.mu.Unlock()
		return
	}
	changed := false
	for family, reset := range observed {
		key := antigravityQuotaPrefix + family
		previous := auth.ModelStates[key]
		if previous != nil && previous.UpdatedAt.After(observedAt) {
			continue
		}
		state := ensureModelState(auth, key)
		state.UpdatedAt = observedAt
		state.Status = StatusActive
		state.Unavailable = reset.After(observedAt)
		state.NextRetryAfter = reset
		state.Quota = QuotaState{}
		state.StatusMessage = ""
		state.LastError = nil
		if state.Unavailable {
			state.Status = StatusError
			state.StatusMessage = "Antigravity quota exhausted until " + reset.UTC().Format(time.RFC3339)
			state.Quota = QuotaState{Exceeded: true, Reason: "quota_snapshot", NextRecoverAt: reset}
		}
		changed = true
	}
	if changed {
		updateAggregatedAvailability(auth, observedAt)
	}
	snapshot := auth.Clone()
	m.mu.Unlock()
	if changed && m.scheduler != nil {
		m.scheduler.upsertAuth(snapshot)
	}
	if changed {
		m.persistCooldownStates(ctx)
	}
}

// RunQuotaRefresh warms quota state on startup and refreshes it independently of
// the desktop quota page. At most four network calls run concurrently; a new
// pass never overlaps an unfinished pass, and service cancellation stops them.
func (m *Manager) RunQuotaRefresh(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		m.refreshQuotaSummaries(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (m *Manager) refreshQuotaSummaries(ctx context.Context) {
	var workers sync.WaitGroup
	slots := make(chan struct{}, 4)
	for _, auth := range m.List() {
		if auth.Provider != "antigravity" || auth.Disabled || auth.Status == StatusDisabled {
			continue
		}
		m.mu.RLock()
		fetcher, ok := m.executors[auth.Provider].(QuotaSummaryFetcher)
		m.mu.RUnlock()
		if !ok {
			continue
		}
		select {
		case <-ctx.Done():
			workers.Wait()
			return
		case slots <- struct{}{}:
		}
		workers.Add(1)
		go func(auth *Auth, fetcher QuotaSummaryFetcher) {
			defer workers.Done()
			defer func() { <-slots }()
			observedAt := time.Now()
			body, err := fetcher.FetchQuotaSummary(ctx, auth)
			if err == nil && ctx.Err() == nil {
				m.ApplyAntigravityQuotaSummary(ctx, auth.ID, body, observedAt)
			}
		}(auth, fetcher)
	}
	workers.Wait()
}
