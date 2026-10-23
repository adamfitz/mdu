// mangadex auth
package mangasrc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type AuthConfig struct {
	SessionToken string    `json:"session_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Username     string    `json:"username,omitempty"`
	Password     string    `json:"password,omitempty"`
	SavedAt      time.Time `json:"saved_at,omitempty"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Result string `json:"result"`
	Token  struct {
		Session string `json:"session"`
		Refresh string `json:"refresh"`
	} `json:"token"`
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "mdu")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func getAuthConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mangadex_auth.json"), nil
}

func LoadAuthConfig() (*AuthConfig, error) {
	path, err := getAuthConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &AuthConfig{}, nil
		}
		return nil, err
	}
	var cfg AuthConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveAuthConfig(cfg *AuthConfig) error {
	path, err := getAuthConfigPath()
	if err != nil {
		return err
	}
	cfg.SavedAt = time.Now()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func Login(username, password string) (*AuthConfig, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password required")
	}
	reqBody := loginRequest{Username: username, Password: password}
	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", mangadexApiBaseUri+"/auth/login", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed: status %d: %s", resp.StatusCode, string(respBody))
	}
	var lr loginResponse
	if err := json.Unmarshal(respBody, &lr); err != nil {
		return nil, err
	}
	cfg := &AuthConfig{
		SessionToken: lr.Token.Session,
		RefreshToken: lr.Token.Refresh,
		Username:     username,
		Password:     password,
		SavedAt:      time.Now(),
	}
	if err := SaveAuthConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Logout() error {
	path, err := getAuthConfigPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (c *AuthConfig) AddAuthHeader(req *http.Request) {
	if c != nil && c.SessionToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.SessionToken)
	}
}
