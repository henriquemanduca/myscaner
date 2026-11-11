package openrouter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"my-scaner/pkg/llm"
)

const (
	// DefaultBaseURL é o endpoint base padrão da API do OpenRouter.
	DefaultBaseURL = "https://openrouter.ai/api/v1"
	// DefaultModel é o modelo multimodal padrão para OCR e visão.
	DefaultModel = "google/gemini-2.0-flash-001"
)

// Extractor é um alias para a interface comum de extração em pkg/llm.
type Extractor = llm.Extractor

// Client gerencia a comunicação com a API do OpenRouter reutilizando a infraestrutura comum.
type Client struct {
	inner *llm.OpenAICompatibleClient
}

// ClientOption permite customizar a criação do cliente.
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// WithBaseURL define uma URL base alternativa.
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) {
		if url != "" {
			c.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// WithModel define o modelo a ser utilizado na requisição.
func WithModel(model string) ClientOption {
	return func(c *clientConfig) {
		if model != "" {
			c.model = model
		}
	}
}

// WithHTTPClient define o cliente HTTP customizado.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *clientConfig) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// NewClient cria e valida uma nova instância do serviço OpenRouter.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("a chave de API do OpenRouter (api_key) é obrigatória")
	}

	cfg := &clientConfig{
		baseURL: DefaultBaseURL,
		model:   DefaultModel,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	inner := llm.NewOpenAICompatibleClient(llm.Config{
		ProviderName: "OpenRouter",
		BaseURL:      cfg.baseURL,
		APIKey:       apiKey,
		Model:        cfg.model,
		HTTPClient:   cfg.httpClient,
		ExtraHeaders: map[string]string{
			"X-Title": "my-scaner",
		},
	})

	return &Client{inner: inner}, nil
}

// ExtractText envia a imagem codificada em Base64 para o OpenRouter e retorna o texto em formato Markdown.
func (c *Client) ExtractText(ctx context.Context, imagePath string) (string, error) {
	return c.inner.ExtractText(ctx, imagePath)
}
