package handlers

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func GetHandler(key, secret, username, url, v, path string, privateKey *rsa.PrivateKey, l *zap.Logger) (*Handler, error) {

	handler := &Handler{
		clientKey:    key,
		clientSecret: secret,
		username:     username,
		privateKey:   privateKey,
		instanceURL:  url,
		authURL:      url + "/services/oauth2/authorize",
		tokenURL:     url + "/services/oauth2/token",
		sobjectsURL:  url + "/services/data/v58.0/sobjects",
		UserAgent:    "sfdataapp (https://github.com/AmitSuresh/sfdataapp, v" + v + ")",
		l:            l,
	}

	jwtTok, err := handler.createJWT(path)
	if err != nil {
		l.Fatal("error creating jwtToken", zap.Error(err))
		return nil, err
	}

	handler.jwtToken = jwtTok
	return handler, nil
}

func (h *Handler) createJWT(p string) (string, error) {

	keyData, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return "", err
	}
	now := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": h.clientKey,
		"sub": h.username,
		"aud": "https://test.salesforce.com",
		"exp": now + 300,
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (h *Handler) GetAccessToken() error {

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("assertion", h.jwtToken)

	req, err := http.NewRequest("POST", h.tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", h.UserAgent)
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}

	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return err
	}

	accessToken, ok := responseData["access_token"].(string)
	if !ok {
		return errors.New("access_token is not a string")
	}
	h.accessToken = accessToken

	refreshToken, ok := responseData["refresh_token"].(string)
	if !ok {
		return errors.New("refresh_token is not a string")
	}
	h.refreshToken = refreshToken

	return err
}

func (h *Handler) BuildDynamicMapping(objectAPI string) (map[string]string, error) {

	metadataURL := fmt.Sprintf("%s/services/data/v53.0/sobjects/%s/describe/", h.instanceURL, objectAPI)

	req, _ := http.NewRequest("GET", metadataURL, nil)
	req.Header.Add("Authorization", "Bearer "+h.accessToken)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	var metadata MetadataResponse

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return nil, err
	}

	mapping := make(FieldAPILabelMapping)
	for _, field := range metadata.Fields {
		mapping[field.Label] = field.Name
	}

	return mapping, nil
}
