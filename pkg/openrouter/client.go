package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultBaseURL é o endpoint base padrão da API do OpenRouter.
	DefaultBaseURL = "https://openrouter.ai/api/v1"
	// DefaultModel é o modelo multimodal padrão para OCR e visão.
	DefaultModel = "google/gemini-2.0-flash-001"
)

// Extractor define a interface para extração de texto de imagens.
type Extractor interface {
	ExtractText(ctx context.Context, imagePath string) (string, error)
}

// Client gerencia a comunicação com a API do OpenRouter.
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// ClientOption permite customizar a criação do cliente.
type ClientOption func(*Client)

// WithBaseURL define uma URL base alternativa.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		if url != "" {
			c.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// WithModel define o modelo a ser utilizado na requisição.
func WithModel(model string) ClientOption {
	return func(c *Client) {
		if model != "" {
			c.model = model
		}
	}
}

// WithHTTPClient define o cliente HTTP customizado.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
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

	c := &Client{
		apiKey:  apiKey,
		baseURL: DefaultBaseURL,
		model:   DefaultModel,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string        `json:"role"`
	Content []messagePart `json:"content"`
}

type messagePart struct {
	Type     string         `json:"type"`
	Text     string         `json:"text,omitempty"`
	ImageURL *imageURLField `json:"image_url,omitempty"`
}

type imageURLField struct {
	URL string `json:"url"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    any    `json:"code"`
	} `json:"error,omitempty"`
}

// ExtractText envia a imagem codificada em Base64 para o OpenRouter e retorna o texto em formato Markdown.
func (c *Client) ExtractText(ctx context.Context, imagePath string) (string, error) {
	dataURL, err := EncodeImageToBase64DataURL(imagePath)
	if err != nil {
		return "", fmt.Errorf("erro no processamento da imagem: %w", err)
	}

	prompt := "Você é um especialista em OCR e digitalização de documentos. " +
		"Analise minuciosamente a imagem fornecida e extraia TODO o texto contido nela fielmente, " +
		"formatando o resultado exclusivamente em Markdown (.md). Mantenha tabelas, listas, títulos e formatações " +
		"equivalentes. Não adicione conversações, introduções ou meta-comentários fora do conteúdo extraído."

	reqBody := chatCompletionRequest{
		Model: c.model,
		Messages: []chatMessage{
			{
				Role: "user",
				Content: []messagePart{
					{
						Type: "text",
						Text: prompt,
					},
					{
						Type: "image_url",
						ImageURL: &imageURLField{
							URL: dataURL,
						},
					},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("falha ao serializar payload da requisição: %w", err)
	}

	endpoint := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("falha ao criar requisição HTTP: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Title", "my-scaner")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("falha na chamada HTTP para o OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("falha ao ler resposta da API: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr chatCompletionResponse
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != nil && apiErr.Error.Message != "" {
			return "", fmt.Errorf("erro da API OpenRouter (status %d): %s", resp.StatusCode, apiErr.Error.Message)
		}
		return "", fmt.Errorf("erro da API OpenRouter (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("falha ao deserializar resposta JSON do OpenRouter: %w", err)
	}

	if len(chatResp.Choices) == 0 || chatResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("a API do OpenRouter retornou resposta vazia")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	return sanitizeMarkdownOutput(content), nil
}

// sanitizeMarkdownOutput remove blocos externos ```markdown ... ``` caso o modelo os tenha adicionado.
func sanitizeMarkdownOutput(content string) string {
	re := regexp.MustCompile("(?s)^```(?:markdown|md)?\\r?\\n(.*?)\\r?\\n```$")
	if matches := re.FindStringSubmatch(content); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return content
}
