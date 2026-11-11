package pdf_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"my-scaner/pkg/pdf"
)

func TestNewPopplerConverter(t *testing.T) {
	conv, err := pdf.NewPopplerConverter()
	if err != nil {
		t.Skipf("pdftoppm não está disponível neste ambiente: %v", err)
	}
	if conv == nil {
		t.Fatal("esperava instância de PopplerConverter, obteve nil")
	}
}

func TestConvertToImages_FileNotFound(t *testing.T) {
	conv, err := pdf.NewPopplerConverter()
	if err != nil {
		t.Skipf("pdftoppm não disponível: %v", err)
	}

	tempDir := t.TempDir()
	_, err = conv.ConvertToImages(context.Background(), "/caminho/inexistente.pdf", tempDir, 150)
	if err == nil {
		t.Fatal("esperava erro para arquivo inexistente, obteve nil")
	}
}

func TestConvertToImages_InvalidPDF(t *testing.T) {
	conv, err := pdf.NewPopplerConverter()
	if err != nil {
		t.Skipf("pdftoppm não disponível: %v", err)
	}

	tempDir := t.TempDir()
	dummyFile := filepath.Join(tempDir, "invalid.pdf")
	if err := os.WriteFile(dummyFile, []byte("isto não é um pdf válido"), 0644); err != nil {
		t.Fatalf("falha ao criar arquivo de teste: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	_, err = conv.ConvertToImages(context.Background(), dummyFile, outputDir, 150)
	if err == nil {
		t.Fatal("esperava erro ao tentar converter PDF inválido, obteve nil")
	}
}
