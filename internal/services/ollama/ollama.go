package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cms-plazareal/internal/config"
)

type Client struct {
	baseURL string
	model   string
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	System string `json:"system,omitempty"`
}

type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.OllamaURL,
		model:   cfg.OllamaModel,
	}
}

// Generate produces a response from Ollama
func (c *Client) Generate(prompt, systemPrompt string) (string, error) {
	reqBody := GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
		System: systemPrompt,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.baseURL+"/api/generate", "application/json", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result GenerateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("ollama response parse error: %w", err)
	}

	return result.Response, nil
}

// ProcessWhatsAppResponse analyzes inbound WhatsApp and generates a smart reply
func (c *Client) ProcessWhatsAppResponse(inboundMessage, eventContext string) (string, error) {
	system := `Eres el asistente virtual del Colegio San Lorenzo de Copiapó.
Responde en español chileno de forma breve, amable y profesional.
Solo responde preguntas relacionadas con el colegio y sus eventos.
Si la pregunta no es relevante, sugiere contactar la secretaría.`

	prompt := fmt.Sprintf(
		"Contexto del evento: %s\n\nMensaje del apoderado: %s\n\nGenera una respuesta apropiada:",
		eventContext, inboundMessage,
	)

	return c.Generate(prompt, system)
}
