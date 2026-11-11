package llm

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

	"my-scaner/pkg/fileutil"
)

// DefaultPrompt é a instrução padrão de sistema enviada para modelos multimodais de OCR.
const DefaultPrompt = "Você é um especialista em OCR e digitalização de documentos. " +
	"Analise minuciosamente a imagem fornecida e extraia TODO o texto contido nela fielmente, " +
	"formatando o resultado exclusivamente em Markdown (.md). Mantenha tabelas, listas, títulos e formatações " +
	"equivalentes. Não adicione conversações, introduções ou meta-comentários fora do conteúdo extraído."

// OpenAICompatibleClient encapsula a lógica de requisição HTTP para endpoints compatíveis com a API OpenAI (OpenRouter, LM Studio, vLLM, etc).
type OpenAICompatibleClient struct {
	ProviderName string
	BaseURL      string
	APIKey       string
	Model        string
	HTTPClient   *http.Client
	ExtraHeaders map[string]string
	Prompt       string
}

// Config contém parâmetros de inicialização do cliente compatível com OpenAI.
type Config struct {
	ProviderName string
	BaseURL      string
	APIKey       string
	Model        string
	HTTPClient   *http.Client
	ExtraHeaders map[string]string
	Prompt       string
	Timeout      time.Duration
}

// NewOpenAICompatibleClient cria e configura uma nova instância de cliente.
func NewOpenAICompatibleClient(cfg Config) *OpenAICompatibleClient {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	prompt := cfg.Prompt
	if prompt == "" {
		prompt = DefaultPrompt
	}

	return &OpenAICompatibleClient{
		ProviderName: cfg.ProviderName,
		BaseURL:      strings.TrimRight(cfg.BaseURL, "/"),
		APIKey:       cfg.APIKey,
		Model:        cfg.Model,
		HTTPClient:   httpClient,
		ExtraHeaders: cfg.ExtraHeaders,
		Prompt:       prompt,
	}
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

// ExtractText executa a extração multimodal de texto codificando a imagem em base64 e chamando o endpoint /chat/completions.
func (c *OpenAICompatibleClient) ExtractText(ctx context.Context, imagePath string) (string, error) {
	dataURL, err := fileutil.EncodeImageToBase64DataURL(imagePath)
	if err != nil {
		return "", fmt.Errorf("erro no processamento da imagem: %w", err)
	}

	reqBody := chatCompletionRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{
				Role: "user",
				Content: []messagePart{
					{
						Type: "text",
						Text: c.Prompt,
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

	endpoint := fmt.Sprintf("%s/chat/completions", c.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("falha ao criar requisição HTTP: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}
	for k, v := range c.ExtraHeaders {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("falha na chamada HTTP para o provedor '%s' (%s): %w", c.ProviderName, endpoint, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("falha ao ler resposta da API de '%s': %w", c.ProviderName, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr chatCompletionResponse
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != nil && apiErr.Error.Message != "" {
			return "", fmt.Errorf("erro da API %s (status %d): %s", c.ProviderName, resp.StatusCode, apiErr.Error.Message)
		}
		return "", fmt.Errorf("erro da API %s (status %d): %s", c.ProviderName, resp.StatusCode, string(respBody))
	}

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("falha ao deserializar resposta JSON de '%s': %w", c.ProviderName, err)
	}

	if len(chatResp.Choices) == 0 || chatResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("a API do provedor '%s' retornou resposta vazia", c.ProviderName)
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	return SanitizeMarkdownOutput(content), nil
}

// SanitizeMarkdownOutput remove blocos externos ```markdown ... ``` caso o modelo os tenha adicionado.
func SanitizeMarkdownOutput(content string) string {
	re := regexp.MustCompile("(?s)^```(?:markdown|md)?\\r?\\n(.*?)\\r?\\n```$")
	if matches := re.FindStringSubmatch(content); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return content
}
