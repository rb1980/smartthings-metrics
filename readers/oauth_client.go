package readers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	SmartThingsTokenURL = "https://auth-global.api.smartthings.com/oauth/token"
)

type OAuthClient struct {
	clientID     string
	clientSecret string
	token        *OAuthToken
	httpClient   *http.Client
}

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	expiresAt    time.Time
}

func NewOAuthClient(clientID, clientSecret string) *OAuthClient {
	return &OAuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *OAuthClient) GetToken() (string, error) {
	if c.token == nil || time.Now().After(c.token.expiresAt) {
		if err := c.refreshToken(); err != nil {
			return "", fmt.Errorf("failed to refresh token: %v", err)
		}
	}
	return c.token.AccessToken, nil
}

func (c *OAuthClient) refreshToken() error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("scope", "devices:read")

	resp, err := c.httpClient.PostForm(SmartThingsTokenURL, data)
	if err != nil {
		return fmt.Errorf("failed to request token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return fmt.Errorf("failed to decode token response: %v", err)
	}

	token.expiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	c.token = &token
	return nil
} 