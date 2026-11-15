package lmstudio_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"scaner-cli/pkg/lmstudio"
)

func TestLMStudio_ExtractText_Success(t *testing.T) {
	tempDir := t.TempDir()
	imgPath := filepath.Join(tempDir, "page.png")
	if err := os.WriteFile(imgPath, []byte("fake-png-data"), 0644); err != nil {
		t.Fatalf("falha ao criar arquivo temporário: %v", err)
	}

	expectedMD := "### Título Local\n\nConteúdo extraído via LM Studio"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("esperava POST, obteve %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("caminho inesperado: %s", r.URL.Path)
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": "```md\n" + expectedMD + "\n```",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := lmstudio.NewClient(
		lmstudio.WithBaseURL(server.URL+"/v1"),
		lmstudio.WithModel("qwen2.5-vl-7b-instruct"),
		lmstudio.WithHTTPClient(server.Client()),
		lmstudio.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("erro ao criar cliente LM Studio: %v", err)
	}

	result, err := client.ExtractText(context.Background(), imgPath)
	if err != nil {
		t.Fatalf("erro inesperado na extração via LM Studio: %v", err)
	}

	if result != expectedMD {
		t.Errorf("resultado diferente do esperado.\nEsperado:\n%s\nObtido:\n%s", expectedMD, result)
	}
}

func TestLMStudio_ExtractText_ServerError(t *testing.T) {
	tempDir := t.TempDir()
	imgPath := filepath.Join(tempDir, "page.png")
	if err := os.WriteFile(imgPath, []byte("fake-png-data"), 0644); err != nil {
		t.Fatalf("falha ao criar arquivo: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"message": "Model out of memory or not loaded",
			},
		})
	}))
	defer server.Close()

	client, err := lmstudio.NewClient(
		lmstudio.WithBaseURL(server.URL+"/v1"),
		lmstudio.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("erro ao criar cliente LM Studio: %v", err)
	}

	_, err = client.ExtractText(context.Background(), imgPath)
	if err == nil {
		t.Fatal("esperava erro ao receber 500 do LM Studio, obteve nil")
	}
}
