package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
)

func quotaSummaryFixture(weekly, short time.Time) []byte {
	return []byte(fmt.Sprintf(`{"groups":[{"buckets":[
	{"bucketId":"gemini-weekly","remainingFraction":0,"resetTime":%q},
	{"bucketId":"gemini-5h","remainingFraction":0,"resetTime":%q}]},
	{"buckets":[{"bucketId":"3p-weekly","remainingFraction":0.6},{"bucketId":"3p-5h","remainingFraction":1}]}]}`, weekly.Format(time.RFC3339Nano), short.Format(time.RFC3339Nano)))
}

func TestAntigravityQuotaSummaryWindowsAndUnknowns(t *testing.T) {
	now := time.Now()
	week, short := now.Add(24*time.Hour), now.Add(time.Hour)
	got := ParseAntigravityQuotaSummary(quotaSummaryFixture(week, short), now)
	if !got["gemini"].Equal(week) || !got["3p"].IsZero() {
		t.Fatalf("unexpected windows: %v", got)
	}
	for _, body := range []string{
		`{}`, `invalid`, `{"groups":[{"buckets":[{"bucketId":"gemini-weekly"}]}]}`,
		`{"groups":[{"buckets":[{"bucketId":"gemini-weekly","remainingFraction":0,"resetTime":"bad"}]}]}`,
		`{"groups":[{"buckets":[{"bucketId":"unknown-weekly","remainingFraction":0,"resetTime":"2099-01-01T00:00:00Z"}]}]}`,
		`{"groups":[{"buckets":[{"bucketId":"gemini-weekly","remainingFraction":0,"resetTime":"2099-01-01T00:00:00Z","disabled":true}]}]}`,
	} {
		if len(ParseAntigravityQuotaSummary([]byte(body), now)) != 0 {
			t.Fatalf("unknown or disabled data changed routing: %s", body)
		}
	}
	positive := ParseAntigravityQuotaSummary([]byte(`{"groups":[{"buckets":[{"bucketId":"gemini-weekly","remainingFraction":0.00001},{"bucketId":"gemini-5h","disabled":true}]}]}`), now)
	if reset, ok := positive["gemini"]; !ok || !reset.IsZero() {
		t.Fatal("rounded UI zero must not mean exhausted")
	}
}

func TestAntigravityQuotaSchedulerSkipsOnlyAffectedModelsAndRecovers(t *testing.T) {
	ctx := context.Background()
	m := NewManager(nil, &FillFirstSelector{}, nil)
	m.RegisterExecutor(schedulerTestExecutor{provider: "antigravity"})
	for _, id := range []string{"quota-a", "quota-b"} {
		registry.GetGlobalRegistry().RegisterClient(id, "antigravity", []*registry.ModelInfo{{ID: "gemini-test"}, {ID: "claude-test"}})
		t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(id) })
		if _, err := m.Register(ctx, &Auth{ID: id, Provider: "antigravity", Status: StatusActive}); err != nil {
			t.Fatal(err)
		}
	}
	// Build the fast scheduler shard before applying observations.
	if _, _, err := m.pickNext(ctx, "antigravity", "gemini-test", cliproxyexecutor.Options{}, nil); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	reset := now.Add(time.Hour)
	stale := m.List()[0]
	for _, a := range m.List() {
		if a.ID == "quota-a" {
			stale = a
		}
	}
	m.ApplyAntigravityQuotaSummary(ctx, "quota-a", quotaSummaryFixture(reset, now.Add(time.Minute)), now)
	for i := 0; i < 4; i++ {
		chosen, _, err := m.pickNext(ctx, "antigravity", "gemini-test", cliproxyexecutor.Options{}, nil)
		if err != nil || chosen.ID != "quota-b" {
			t.Fatalf("exhausted account selected: %v %v", chosen, err)
		}
	}
	chosen, _, err := m.pickNext(ctx, "antigravity", "claude-test", cliproxyexecutor.Options{}, nil)
	if err != nil || chosen.ID != "quota-a" {
		t.Fatalf("healthy family incorrectly blocked: %v %v", chosen, err)
	}
	// Neither a stale token update nor an in-flight successful inference clears it.
	if _, err = m.Update(ctx, stale); err != nil {
		t.Fatal(err)
	}
	m.MarkResult(ctx, Result{AuthID: "quota-a", Model: "gemini-test", Success: true})
	for _, a := range m.List() {
		if a.ID == "quota-a" {
			blocked, _, until := isAuthBlockedForModel(a, "gemini-test(high)", now)
			if !blocked || !until.Equal(reset) {
				t.Fatalf("quota lost after refresh/success: %v %v", blocked, until)
			}
			if blocked, _, _ := isAuthBlockedForModel(a, "gemini-test", reset.Add(time.Nanosecond)); blocked {
				t.Fatal("still blocked after reset")
			}
		}
	}
	// A valid newer positive observation clears only its own group constraint.
	partial := []byte(`{"groups":[{"buckets":[{"bucketId":"gemini-5h","remainingFraction":1}]}]}`)
	m.ApplyAntigravityQuotaSummary(ctx, "quota-a", partial, now.Add(time.Second))
	chosen, _, err = m.pickNext(ctx, "antigravity", "gemini-test", cliproxyexecutor.Options{}, nil)
	if err != nil || chosen.ID != "quota-b" {
		t.Fatalf("partial snapshot erased weekly exhaustion: %v %v", chosen, err)
	}
	positive := []byte(`{"groups":[{"buckets":[{"bucketId":"gemini-weekly","remainingFraction":1},{"bucketId":"gemini-5h","remainingFraction":1}]}]}`)
	m.ApplyAntigravityQuotaSummary(ctx, "quota-a", positive, now.Add(time.Second))
	m.ApplyAntigravityQuotaSummary(ctx, "quota-a", quotaSummaryFixture(reset, reset), now)
	chosen, _, err = m.pickNext(ctx, "antigravity", "gemini-test", cliproxyexecutor.Options{}, nil)
	if err != nil || chosen.ID != "quota-a" {
		t.Fatalf("recovered account not selected / stale response won: %v %v", chosen, err)
	}
}

type quotaPollingExecutor struct {
	schedulerTestExecutor
	payload []byte
}

func (e quotaPollingExecutor) FetchQuotaSummary(context.Context, *Auth) ([]byte, error) {
	return e.payload, nil
}

func TestAntigravityQuotaPollingUpdatesWithoutUI(t *testing.T) {
	ctx := context.Background()
	m := NewManager(nil, &RoundRobinSelector{}, nil)
	m.RegisterExecutor(quotaPollingExecutor{schedulerTestExecutor{provider: "antigravity"}, quotaSummaryFixture(time.Now().Add(time.Hour), time.Now().Add(time.Minute))})
	if _, err := m.Register(ctx, &Auth{ID: "polling-quota", Provider: "antigravity"}); err != nil {
		t.Fatal(err)
	}
	m.refreshQuotaSummaries(ctx)
	if blocked, _, _ := isAuthBlockedForModel(m.List()[0], "gemini-test", time.Now()); !blocked {
		t.Fatal("background quota not applied")
	}
}

func TestAntigravityQuotaRestoresAndSchedulerReopensAtReset(t *testing.T) {
	ctx := context.Background()
	m := NewManager(nil, &FillFirstSelector{}, nil)
	if _, err := m.Register(ctx, &Auth{ID: "quota-reset", Provider: "antigravity"}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	reset := now.Add(150 * time.Millisecond)
	m.ApplyAntigravityQuotaSummary(ctx, "quota-reset", quotaSummaryFixture(reset, reset), now)
	restored := NewManager(nil, &FillFirstSelector{}, nil)
	restored.RegisterExecutor(schedulerTestExecutor{provider: "antigravity"})
	registry.GetGlobalRegistry().RegisterClient("quota-reset", "antigravity", []*registry.ModelInfo{{ID: "gemini-test"}, {ID: "claude-test"}})
	t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient("quota-reset") })
	if _, err := restored.Register(ctx, &Auth{ID: "quota-reset", Provider: "antigravity"}); err != nil {
		t.Fatal(err)
	}
	restored.SetCooldownStateStore(&mockCooldownStateStore{records: m.cooldownStateRecordsSnapshot()})
	if err := restored.RestoreCooldownStates(ctx); err != nil {
		t.Fatal(err)
	}
	if _, _, err := restored.pickNext(ctx, "antigravity", "gemini-test", cliproxyexecutor.Options{}, nil); err == nil {
		t.Fatal("restored exhausted model was selected")
	}
	if _, _, err := restored.pickNext(ctx, "antigravity", "claude-test", cliproxyexecutor.Options{}, nil); err != nil {
		t.Fatalf("restored state blocks unrelated family: %v", err)
	}
	time.Sleep(time.Until(reset) + 5*time.Millisecond)
	if _, _, err := restored.pickNext(ctx, "antigravity", "gemini-test", cliproxyexecutor.Options{}, nil); err != nil {
		t.Fatalf("scheduler did not reopen at quota reset: %v", err)
	}
}
