package openrouter

import "my-scaner/pkg/fileutil"

// SupportedImageExtensions reexporta as extensões suportadas de fileutil para manter compatibilidade.
var SupportedImageExtensions = fileutil.SupportedImageExtensions

// IsPDF reexporta a função de verificação de PDF.
var IsPDF = fileutil.IsPDF

// ValidateInputFile reexporta a validação de arquivo de entrada.
var ValidateInputFile = fileutil.ValidateInputFile

// ValidateImageFile reexporta a validação de imagem.
var ValidateImageFile = fileutil.ValidateImageFile

// EncodeImageToBase64DataURL reexporta a codificação de imagem em Data URL.
var EncodeImageToBase64DataURL = fileutil.EncodeImageToBase64DataURL

// OutputMarkdownPath reexporta o cálculo do caminho de saída markdown.
var OutputMarkdownPath = fileutil.OutputMarkdownPath
