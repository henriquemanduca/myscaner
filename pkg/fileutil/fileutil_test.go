package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
	return path
}

func TestValidateInputFile(t *testing.T) {
	dir := t.TempDir()
	pdf := writeTemp(t, "doc.pdf", []byte("%PDF-1.7"))
	png := writeTemp(t, "img.png", []byte("png"))
	txt := writeTemp(t, "notes.txt", []byte("text"))

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "pdf", path: pdf},
		{name: "image", path: png},
		{name: "missing", path: filepath.Join(dir, "missing.pdf"), wantErr: "não encontrado"},
		{name: "directory", path: dir, wantErr: "diretório"},
		{name: "unsupported", path: txt, wantErr: "não suportado"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputFile(tt.path)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestIsPDF(t *testing.T) {
	if !IsPDF("report.PDF") {
		t.Fatal("expected uppercase .PDF to be detected")
	}
	if IsPDF("report.png") {
		t.Fatal("expected .png not to be a PDF")
	}
}

func TestOutputMarkdownPath(t *testing.T) {
	if got, want := OutputMarkdownPath("/tmp/report.pdf"), "/tmp/report.md"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := OutputMarkdownPath("/tmp/page.PNG"), "/tmp/page.md"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestEncodeImageToBase64DataURL(t *testing.T) {
	path := writeTemp(t, "pixel.png", []byte{0x89, 'P', 'N', 'G'})

	dataURL, err := EncodeImageToBase64DataURL(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(dataURL, "data:image/png;base64,") {
		t.Fatalf("unexpected data URL prefix: %q", dataURL)
	}
}

func TestEncodeImageRejectsNonImage(t *testing.T) {
	path := writeTemp(t, "notes.txt", []byte("text"))
	if _, err := EncodeImageToBase64DataURL(path); err == nil {
		t.Fatal("expected error for non-image input")
	}
}
