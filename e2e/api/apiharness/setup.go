//go:build api

package apiharness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joho/godotenv"

	"github.com/verzac/grocer-discord-bot/dto"
)

type APITestSession struct {
	client    *http.Client
	baseURL   string
	authValue string
}

func dotEnvPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("apiharness.dotEnvPath: runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".env"))
}

func SetupAPI() *APITestSession {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := godotenv.Load(dotEnvPath()); err != nil {
		log.Println("Cannot load .env file:", err.Error())
	}

	authHdr := stringsTrimQuotes(strings.TrimSpace(os.Getenv("E2E_API_HEADER")))
	if authHdr == "" {
		panic("E2E_API_HEADER is required for API E2E tests.")
	}

	baseURL := strings.TrimSpace(os.Getenv("E2E_API_BASE_URL"))
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &APITestSession{
		client:    http.DefaultClient,
		baseURL:   baseURL,
		authValue: authHdr,
	}
}

func stringsTrimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func (s *APITestSession) applyAuth(req *http.Request) {
	req.Header.Set("Authorization", s.authValue)
}

func (s *APITestSession) Get(path string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	s.applyAuth(req)
	return s.client.Do(req)
}

func (s *APITestSession) PostJSON(path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, s.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	s.applyAuth(req)
	return s.client.Do(req)
}

func (s *APITestSession) DeleteJSON(path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, s.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	s.applyAuth(req)
	return s.client.Do(req)
}

func (s *APITestSession) DeleteNoBody(path string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, s.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	s.applyAuth(req)
	return s.client.Do(req)
}

func ReadBodyAndClose(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func (s *APITestSession) FetchGroceryLists() (*dto.GuildGroceryList, int, error) {
	res, err := s.Get("/grocery-lists")
	if err != nil {
		return nil, 0, err
	}
	b, err := ReadBodyAndClose(res)
	if err != nil {
		return nil, res.StatusCode, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, nil
	}
	var out dto.GuildGroceryList
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, res.StatusCode, err
	}
	return &out, res.StatusCode, nil
}

func (s *APITestSession) CleanupAllGroceries() error {
	lists, status, err := s.FetchGroceryLists()
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("cleanup: GET /grocery-lists returned %d", status)
	}
	ids := make([]uint, 0, len(lists.GroceryEntries))
	for _, e := range lists.GroceryEntries {
		ids = append(ids, e.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	for start := 0; start < len(ids); start += 300 {
		end := start + 300
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]
		payload, err := json.Marshal(dto.GroceryBatchDeleteRequest{IDs: chunk})
		if err != nil {
			return err
		}
		res, err := s.DeleteJSON("/groceries", payload)
		if err != nil {
			return err
		}
		_, _ = ReadBodyAndClose(res)
		if res.StatusCode != http.StatusNoContent {
			return fmt.Errorf("batch delete groceries: unexpected status %d", res.StatusCode)
		}
	}
	return nil
}

