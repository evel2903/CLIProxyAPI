package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

// FetchQuotaSummary uses the credential's normal token refresh and proxy
// transport. It never sends an inference request or uses paid credits.
func (e *AntigravityExecutor) FetchQuotaSummary(ctx context.Context, auth *cliproxyauth.Auth) ([]byte, error) {
	token, updated, err := e.ensureAccessToken(ctx, auth)
	if err != nil {
		return nil, err
	}
	if updated != nil {
		auth = updated
	}
	project := antigravityProjectIDFromAuth(auth)
	if project == "" {
		return nil, fmt.Errorf("Antigravity quota: missing project")
	}
	body, err := json.Marshal(map[string]string{"project": project})
	if err != nil {
		return nil, err
	}
	baseURLs := antigravityBaseURLFallbackOrder(auth)
	if len(baseURLs) == 0 {
		return nil, fmt.Errorf("Antigravity quota: missing endpoint")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURLs[0], "/")+"/v1internal:retrieveUserQuotaSummary", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", resolveUserAgent(auth))
	response, err := newAntigravityHTTPClient(ctx, e.cfg, auth, 0).Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if errClose := response.Body.Close(); errClose != nil {
			log.WithError(errClose).Debug("Antigravity quota response close failed")
		}
	}()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Antigravity quota: HTTP %d", response.StatusCode)
	}
	const limit = 2 << 20
	payload, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if len(payload) > limit {
		return nil, fmt.Errorf("Antigravity quota response too large")
	}
	return payload, nil
}
