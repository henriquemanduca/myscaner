package lmstudio

import (
	"context"
	"net/http"
	"strings"
	"time"

	"my-scaner/pkg/llm"
)

const (
	// DefaultBaseURL é o endpoint padrão da API local do LM Studio.
	DefaultBaseURL = "http://localhost:1234/v1"
	// DefaultModel é o valor padrão para modelo no LM Studio.
	// Quando não especificado, modelos locais do LM Studio podem aceitar o modelo carregado atualmente ou identificador customizado.
	DefaultModel = "default"
)

// Extractor é um alias para a interface comum de extração em pkg/llm.
type Extractor = llm.Extractor

// Client gerencia a comunicação com a API local do LM Studio.
type Client struct {
	inner *llm.OpenAICompatibleClient
}

// ClientOption permite customizar a criação do cliente LM Studio.
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL    string
	model      string
	apiKey     string
	httpClient *http.Client
	timeout    time.Duration
}

// WithBaseURL define a URL base da API do LM Studio (ex: http://localhost:1234/v1).
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) {
		if url != "" {
			c.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// WithModel define o identificador do modelo carregado no LM Studio.
func WithModel(model string) ClientOption {
	return func(c *clientConfig) {
		if model != "" {
			c.model = model
		}
	}
}

// WithAPIKey define a chave de API (opcional no LM Studio local).
func WithAPIKey(key string) ClientOption {
	return func(c *clientConfig) {
		c.apiKey = key
	}
}

// WithHTTPClient define um cliente HTTP customizado.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *clientConfig) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithTimeout define o timeout máximo das requisições para inferência local.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		if timeout > 0 {
			c.timeout = timeout
		}
	}
}

// NewClient cria uma nova instância do cliente para o LM Studio local.
func NewClient(opts ...ClientOption) *Client {
	cfg := &clientConfig{
		baseURL: DefaultBaseURL,
		model:   DefaultModel,
		timeout: 180 * time.Second, // Inferência local pode exigir maior tempo de resposta
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{Timeout: cfg.timeout}
	}

	inner := llm.NewOpenAICompatibleClient(llm.Config{
		ProviderName: "LM Studio",
		BaseURL:      cfg.baseURL,
		APIKey:       cfg.apiKey,
		Model:        cfg.model,
		HTTPClient:   cfg.httpClient,
		Timeout:      cfg.timeout,
	})

	return &Client{inner: inner}
}

// ExtractText envia a imagem para o LM Studio local e retorna o texto extraído em Markdown.
func (c *Client) ExtractText(ctx context.Context, imagePath string) (string, error) {
	return c.inner.ExtractText(ctx, imagePath)
}
