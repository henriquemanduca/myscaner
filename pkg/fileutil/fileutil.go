package fileutil

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SupportedImageExtensions lista as extensões de imagens suportadas.
var SupportedImageExtensions = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".tiff": "image/tiff",
	".tif":  "image/tiff",
}

// IsPDF verifica se o arquivo possui extensão .pdf.
func IsPDF(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".pdf"
}

// ValidateInputFile valida se o caminho existe, é um arquivo comum e possui extensão permitida (imagem ou PDF).
func ValidateInputFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("arquivo não encontrado: %s", path)
		}
		return fmt.Errorf("erro ao acessar arquivo %s: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("o caminho informado é um diretório, esperado arquivo de imagem ou PDF: %s", path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".pdf" {
		return nil
	}

	if _, ok := SupportedImageExtensions[ext]; !ok {
		supportedList := make([]string, 0, len(SupportedImageExtensions)+1)
		for k := range SupportedImageExtensions {
			supportedList = append(supportedList, k)
		}
		supportedList = append(supportedList, ".pdf")
		return fmt.Errorf("tipo de arquivo '%s' não suportado. Tipos válidos: %s", ext, strings.Join(supportedList, ", "))
	}

	return nil
}

// ValidateImageFile valida se o caminho existe, é um arquivo comum e possui extensão de imagem permitida.
func ValidateImageFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("arquivo de imagem não encontrado: %s", path)
		}
		return fmt.Errorf("erro ao acessar arquivo %s: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("o caminho informado é um diretório, esperado arquivo de imagem: %s", path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	if _, ok := SupportedImageExtensions[ext]; !ok {
		supportedList := make([]string, 0, len(SupportedImageExtensions))
		for k := range SupportedImageExtensions {
			supportedList = append(supportedList, k)
		}
		return fmt.Errorf("tipo de arquivo '%s' não suportado. Tipos válidos: %s", ext, strings.Join(supportedList, ", "))
	}

	return nil
}

// EncodeImageToBase64DataURL lê a imagem e a codifica no formato data URL (data:<mime>;base64,<dados>).
func EncodeImageToBase64DataURL(path string) (string, error) {
	if err := ValidateImageFile(path); err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("falha ao ler bytes da imagem: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	mimeType, ok := SupportedImageExtensions[ext]
	if !ok || mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

// OutputMarkdownPath calcula o caminho do arquivo .md no mesmo diretório e com o mesmo nome base do arquivo de entrada (imagem ou pdf).
func OutputMarkdownPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	return strings.TrimSuffix(inputPath, ext) + ".md"
}
