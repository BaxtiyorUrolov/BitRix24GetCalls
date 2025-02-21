package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"bitrix/models"
	"bitrix/storage"
)

const oauthURL = "https://oauth.bitrix.info/oauth/token/"

// ExchangeCodeForToken - code → access_token
func ExchangeCodeForToken(db *sql.DB, code, clientID, clientSecret, redirectURI string) (*models.TokenInfo, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("code", code)

	resp, err := http.PostForm(oauthURL, data)
	if err != nil {
		return nil, fmt.Errorf("token so'rovda xatolik: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token so'rov javobi xato status: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken    string `json:"access_token"`
		RefreshToken   string `json:"refresh_token"`
		ExpiresIn      int    `json:"expires_in"`
		Scope          string `json:"scope"`
		Domain         string `json:"domain"`
		ServerEndpoint string `json:"server_endpoint"`
		ClientEndpoint string `json:"client_endpoint"`
		MemberID       string `json:"member_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON parse xatolik: %v", err)
	}

	tokenInfo := &models.TokenInfo{
		PortalDomain:   result.Domain,
		MemberID:       result.MemberID,
		AccessToken:    result.AccessToken,
		RefreshToken:   result.RefreshToken,
		ExpiresIn:      result.ExpiresIn,
		Scope:          result.Scope,
		LastUpdate:     time.Now(),
		ClientEndpoint: result.ClientEndpoint,
	}

	if err := storage.InsertOrUpdateToken(db, tokenInfo); err != nil {
		return nil, fmt.Errorf("token saqlashda xatolik: %v", err)
	}

	return tokenInfo, nil
}

// RefreshToken - token muddati tugasa, yangilash
func RefreshToken(db *sql.DB, t *models.TokenInfo, clientID, clientSecret string) error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", t.RefreshToken)

	resp, err := http.PostForm(oauthURL, data)
	if err != nil {
		return fmt.Errorf("refresh token so'rovda xatolik: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh token so'rov javobi xato status: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken    string `json:"access_token"`
		RefreshToken   string `json:"refresh_token"`
		ExpiresIn      int    `json:"expires_in"`
		Scope          string `json:"scope"`
		Domain         string `json:"domain"`
		ClientEndpoint string `json:"client_endpoint"`
		MemberID       string `json:"member_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("refresh JSON parse xatolik: %v", err)
	}

	t.AccessToken = result.AccessToken
	t.RefreshToken = result.RefreshToken
	t.ExpiresIn = result.ExpiresIn
	t.Scope = result.Scope
	t.LastUpdate = time.Now()
	if result.ClientEndpoint != "" {
		t.ClientEndpoint = result.ClientEndpoint
	}

	if err := storage.InsertOrUpdateToken(db, t); err != nil {
		return fmt.Errorf("token yangilashda xatolik: %v", err)
	}
	return nil
}

// IsTokenExpired - tokenning amal qilish muddatini tekshirish
func IsTokenExpired(t *models.TokenInfo) bool {
	expireTime := t.LastUpdate.Add(time.Duration(t.ExpiresIn) * time.Second)
	return time.Now().After(expireTime)
}
