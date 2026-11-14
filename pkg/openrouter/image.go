package openrouter

import "my-scaner/pkg/fileutil"

// SupportedImageExtensions reexporta as extensões suportadas de fileutil para manter compatibilidade.
func SupportedImageExtensions() map[string]string {
	return fileutil.SupportedImageExtensions()
}

// IsPDF reexporta a função de verificação de PDF.
func IsPDF(path string) bool {
	return fileutil.IsPDF(path)
}

// ValidateInputFile reexporta a validação de arquivo de entrada.
func ValidateInputFile(path string) error {
	return fileutil.ValidateInputFile(path)
}

// ValidateImageFile reexporta a validação de imagem.
func ValidateImageFile(path string) error {
	return fileutil.ValidateImageFile(path)
}

// EncodeImageToBase64DataURL reexporta a codificação de imagem em Data URL.
func EncodeImageToBase64DataURL(path string) (string, error) {
	return fileutil.EncodeImageToBase64DataURL(path)
}

// OutputMarkdownPath reexporta o cálculo do caminho de saída markdown.
func OutputMarkdownPath(inputPath string) string {
	return fileutil.OutputMarkdownPath(inputPath)
}
