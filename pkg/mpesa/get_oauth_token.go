package mpesa

import (
	"encoding/json"
	"net/http"
)

const (
	mpesaBaseURL = "https://sandbox.safaricom.co.ke"
)

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

func GetOAuthToken(consumerKey, consumerSecret string) (*OAuthResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", mpesaBaseURL+"/oauth/v1/generate?grant_type=client_credentials", nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(consumerKey, consumerSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var oauthResponse OAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthResponse); err != nil {
		return nil, err
	}

	return &oauthResponse, nil
}
