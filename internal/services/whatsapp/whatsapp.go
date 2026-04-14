package whatsapp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"jcp-gestioninmobiliaria/internal/config"
)

type Client struct {
	accountSID string
	authToken  string
	fromNumber string
}

type SendResult struct {
	MessageSID string `json:"sid"`
	Status     string `json:"status"`
}

func New(cfg *config.Config) *Client {
	return &Client{
		accountSID: cfg.TwilioAccountSID,
		authToken:  cfg.TwilioAuthToken,
		fromNumber: cfg.TwilioFromNumber,
	}
}

// Send dispatches a WhatsApp message via Twilio API
func (c *Client) Send(to, body string) (*SendResult, error) {
	if c.accountSID == "" || c.authToken == "" {
		return nil, fmt.Errorf("twilio credentials not configured")
	}

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", c.accountSID)

	data := url.Values{}
	data.Set("From", c.fromNumber)
	data.Set("To", "whatsapp:"+to)
	data.Set("Body", body)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.accountSID, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twilio request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("twilio error %d: %s", resp.StatusCode, string(respBody))
	}

	var result SendResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse twilio response: %w", err)
	}

	return &result, nil
}

// SendEventNotification sends a formatted event notification
func (c *Client) SendEventNotification(to, eventTitle, eventDate, eventDesc string) (*SendResult, error) {
	body := fmt.Sprintf(
		"📢 *Colegio San Lorenzo*\n\n*%s*\n📅 %s\n\n%s\n\n_Responde: ✅ Asistiré | ❌ No asistiré | ❓ Tengo dudas_",
		eventTitle, eventDate, eventDesc,
	)
	return c.Send(to, body)
}
