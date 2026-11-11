package openrouter_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"my-scaner/pkg/openrouter"
)

func TestValidateInputFile(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Arquivo de imagem válido (.png)
	validPNG := filepath.Join(tempDir, "test.png")
	if err := os.WriteFile(validPNG, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo temporário: %v", err)
	}
	if err := openrouter.ValidateInputFile(validPNG); err != nil {
		t.Errorf("esperava sucesso para .png válido, obteve: %v", err)
	}

	// 2. Arquivo PDF válido (.pdf)
	validPDF := filepath.Join(tempDir, "test.pdf")
	if err := os.WriteFile(validPDF, []byte("%PDF-1.4 fake"), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo: %v", err)
	}
	if err := openrouter.ValidateInputFile(validPDF); err != nil {
		t.Errorf("esperava sucesso para .pdf válido, obteve: %v", err)
	}
	if !openrouter.IsPDF(validPDF) {
		t.Errorf("esperava IsPDF true para %s", validPDF)
	}

	// 3. Arquivo com extensão inválida (.txt)
	invalidTXT := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(invalidTXT, []byte("plain text"), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo: %v", err)
	}
	if err := openrouter.ValidateInputFile(invalidTXT); err == nil {
		t.Errorf("esperava erro para extensão não suportada (.txt), obteve nil")
	}

	// 4. Arquivo inexistente
	nonExistent := filepath.Join(tempDir, "does-not-exist.pdf")
	if err := openrouter.ValidateInputFile(nonExistent); err == nil {
		t.Errorf("esperava erro para arquivo inexistente, obteve nil")
	}

	// 5. Diretório ao invés de arquivo
	if err := openrouter.ValidateInputFile(tempDir); err == nil {
		t.Errorf("esperava erro ao passar diretório, obteve nil")
	}
}

func TestValidateImageFile(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Arquivo válido (.png)
	validPNG := filepath.Join(tempDir, "test.png")
	if err := os.WriteFile(validPNG, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo temporário: %v", err)
	}
	if err := openrouter.ValidateImageFile(validPNG); err != nil {
		t.Errorf("esperava sucesso para .png válido, obteve: %v", err)
	}

	// 2. Arquivo com extensão inválida (.txt)
	invalidTXT := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(invalidTXT, []byte("plain text"), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo: %v", err)
	}
	if err := openrouter.ValidateImageFile(invalidTXT); err == nil {
		t.Errorf("esperava erro para extensão não suportada (.txt), obteve nil")
	}

	// 3. Arquivo inexistente
	nonExistent := filepath.Join(tempDir, "does-not-exist.jpg")
	if err := openrouter.ValidateImageFile(nonExistent); err == nil {
		t.Errorf("esperava erro para arquivo inexistente, obteve nil")
	}

	// 4. Diretório ao invés de arquivo
	if err := openrouter.ValidateImageFile(tempDir); err == nil {
		t.Errorf("esperava erro ao passar diretório, obteve nil")
	}
}

func TestOutputMarkdownPath(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"/path/to/doc.png", "/path/to/doc.md"},
		{"./image.JPEG", "./image.md"},
		{"invoice.webp", "invoice.md"},
		{"documento.pdf", "documento.md"},
		{"/var/data/relatorio.PDF", "/var/data/relatorio.md"},
	}

	for _, tc := range cases {
		got := openrouter.OutputMarkdownPath(tc.input)
		if got != tc.expected {
			t.Errorf("OutputMarkdownPath(%q) = %q; esperado %q", tc.input, got, tc.expected)
		}
	}
}

func TestExtractText_Success(t *testing.T) {
	tempDir := t.TempDir()
	imgPath := filepath.Join(tempDir, "sample.jpg")
	if err := os.WriteFile(imgPath, []byte("fake-jpeg-bytes"), 0644); err != nil {
		t.Fatalf("falha ao criar arquivo de teste: %v", err)
	}

	expectedMarkdown := "# Documento Extraído\n\n- Item 1\n- Item 2"

	// Mock do servidor OpenRouter
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("esperava método POST, recebeu %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("header Authorization inválido: %s", r.Header.Get("Authorization"))
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": "```markdown\n" + expectedMarkdown + "\n```",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := openrouter.NewClient("test-api-key",
		openrouter.WithBaseURL(server.URL),
		openrouter.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("erro ao criar cliente: %v", err)
	}

	got, err := client.ExtractText(context.Background(), imgPath)
	if err != nil {
		t.Fatalf("falha inesperada na extração: %v", err)
	}

	if got != expectedMarkdown {
		t.Errorf("conteúdo extraído diferente.\nEsperado:\n%s\nObtido:\n%s", expectedMarkdown, got)
	}
}

func TestExtractText_APIError(t *testing.T) {
	tempDir := t.TempDir()
	imgPath := filepath.Join(tempDir, "sample.png")
	if err := os.WriteFile(imgPath, []byte("png-bytes"), 0644); err != nil {
		t.Fatalf("falha ao criar arquivo: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"message": "Invalid API Key",
			},
		})
	}))
	defer server.Close()

	client, err := openrouter.NewClient("bad-key",
		openrouter.WithBaseURL(server.URL),
		openrouter.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("erro ao instanciar cliente: %v", err)
	}

	_, err = client.ExtractText(context.Background(), imgPath)
	if err == nil {
		t.Fatal("esperava erro de autenticação, recebeu nil")
	}
}
