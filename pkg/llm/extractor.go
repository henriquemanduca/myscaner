package llm

import "context"

// Extractor define o contrato comum para extração de texto em Markdown a partir de uma imagem.
type Extractor interface {
	ExtractText(ctx context.Context, imagePath string) (string, error)
}
