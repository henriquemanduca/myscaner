package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "https", url: "https://openrouter.ai/api/v1"},
		{name: "http", url: "http://localhost:1234/v1"},
		{name: "no scheme", url: "localhost:1234", wantErr: true},
		{name: "wrong scheme", url: "ftp://example.com", wantErr: true},
		{name: "no host", url: "https://", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBaseURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateBaseURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeMarkdownOutput(t *testing.T) {
	if got, want := SanitizeMarkdownOutput("```markdown\n# Title\n```"), "# Title"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := SanitizeMarkdownOutput("plain text"), "plain text"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNewOpenAICompatibleClientValidation(t *testing.T) {
	if _, err := NewOpenAICompatibleClient(Config{ProviderName: "x", BaseURL: "", Model: "m"}); err == nil {
		t.Fatal("expected error for empty base URL")
	}
	if _, err := NewOpenAICompatibleClient(Config{ProviderName: "x", BaseURL: "http://localhost:1234/v1", Model: ""}); err == nil {
		t.Fatal("expected error for empty model")
	}
}

func TestExtractText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected Authorization header: %q", r.Header.Get("Authorization"))
		}
		var req chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if req.Model != "test-model" || len(req.Messages) != 1 {
			t.Errorf("unexpected request payload: %+v", req)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"` +
			"```markdown\\n# Extracted\\n```" + `"}}]}`))
	}))
	defer server.Close()

	img := filepath.Join(t.TempDir(), "page.png")
	if err := os.WriteFile(img, []byte{0x89, 'P', 'N', 'G'}, 0o600); err != nil {
		t.Fatalf("writing image: %v", err)
	}

	client, err := NewOpenAICompatibleClient(Config{
		ProviderName: "test",
		BaseURL:      server.URL,
		APIKey:       "test-key",
		Model:        "test-model",
	})
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	got, err := client.ExtractText(context.Background(), img)
	if err != nil {
		t.Fatalf("ExtractText: %v", err)
	}
	if got != "# Extracted" {
		t.Fatalf("got %q, want %q", got, "# Extracted")
	}
}

func TestExtractTextPropagatesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	img := filepath.Join(t.TempDir(), "page.png")
	if err := os.WriteFile(img, []byte{0x89, 'P', 'N', 'G'}, 0o600); err != nil {
		t.Fatalf("writing image: %v", err)
	}

	client, err := NewOpenAICompatibleClient(Config{
		ProviderName: "test",
		BaseURL:      server.URL,
		Model:        "test-model",
	})
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	_, err = client.ExtractText(context.Background(), img)
	if err == nil || !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("expected API error, got %v", err)
	}
}
