package ingredients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/verzac/grocer-discord-bot/config"
)

type n8nRequest struct {
	URL string `json:"url"`
}

type n8nResponse struct {
	List  []string `json:"list,omitempty"`
	Error string   `json:"error,omitempty"`
}

func IsErrFetchRecipeNotFound(err error) bool {
	errMsg := err.Error()
	if strings.Contains(errMsg, "No recipe URL found in video description.") {
		return true
	}
	return false
}

func (s *IngredientsServiceImpl) FetchIngredients(ctx context.Context, url string) ([]string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, errors.New("url is required")
	}
	webhookURL := strings.TrimSpace(config.GetN8NWebhookIngredients())
	if webhookURL == "" {
		return nil, errors.New("N8N_WEBHOOK_INGREDIENTS is not configured")
	}

	body, err := json.Marshal(n8nRequest{URL: url})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.GetN8NApiJWT()))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errObj struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(respBody, &errObj)
		if errObj.Error != "" {
			return nil, fmt.Errorf("%s", errObj.Error)
		}
		return nil, fmt.Errorf("n8n webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}
	var parsed n8nResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("invalid n8n response: %w", err)
	}
	if parsed.Error != "" {
		return nil, fmt.Errorf("%s", parsed.Error)
	}
	out := make([]string, 0, len(parsed.List))
	for _, line := range parsed.List {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out, nil
}
