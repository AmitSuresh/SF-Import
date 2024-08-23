package handlers

import (
	"bytes"
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

	rbmq "github.com/AmitSuresh/sfdataapp/rabbitmq"
)

const VERSION = "0.0.1"

func GetHandler(clientID, secret, username, url, v, path, sfEnv string, rbmqCfg *rbmq.Config, l *zap.Logger) (*Handler, error) {

	handler := &Handler{
		clientID:      clientID,
		clientSecret:  secret,
		username:      username,
		sfEnv:         sfEnv,
		instanceURL:   url,
		authURL:       url + "/services/oauth2/authorize",
		tokenURL:      url + "/services/oauth2/token",
		sobjectsURL:   url + "/services/data/v58.0/sobjects",
		queryURL:      url + "/services/data/v58.0/query?q=",
		uiapiURL:      url + "/services/data/v58.0/ui-api/object-info/",
		uiapibatchURL: url + "/services/data/v58.0/ui-api/records/batch",
		ingestURL:     url + "/services/data/v61.0/jobs/ingest",

		UserAgent: "sfdataapp (https://github.com/AmitSuresh/sfdataapp, v" + VERSION + ")",
		l:         l,
		pKeyPath:  path,
		client:    &http.Client{Timeout: 30 * time.Second},
	}

	jwtTok, err := handler.createJWT(handler.pKeyPath, handler.sfEnv)
	if err != nil {
		l.Fatal("error creating jwtToken", zap.Error(err))
		return nil, err
	}

	handler.jwtToken = jwtTok

	if err := handler.GetAccessToken(); err != nil {
		l.Fatal("error accessing", zap.Error(err))
		return nil, err
	}

	handler.amqpCh, handler.amqpClose, err = rbmq.ConnectAmqp(rbmqCfg, handler.l)
	if err != nil {
		l.Fatal("failed to connect to RabbitMQ", zap.Error(err))
		return nil, err
	}

	return handler, nil
}

func (h *Handler) createJWT(p, s string) (string, error) {

	keyData, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return "", err
	}
	expirationTime := time.Now().Add(30 * time.Minute).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": h.clientID,
		"sub": h.username,
		"aud": fmt.Sprintf("https://%s.salesforce.com", s),
		"exp": expirationTime,
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	h.l.Info("token", zap.Any("", tokenString))
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

	resp, err := h.client.Do(req)
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

	return err
}

func (h *Handler) BuildDynamicMapping(objectAPI string) (map[string]string, error) {

	metadataURL := fmt.Sprintf("%s/services/data/v53.0/sobjects/%s/describe/", h.instanceURL, objectAPI)

	req, _ := http.NewRequest("GET", metadataURL, nil)
	req.Header.Add("Authorization", "Bearer "+h.accessToken)

	resp, err := h.client.Do(req)
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

func ToJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

func FromJSON(i interface{}, r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (h *Handler) handleNewRequest(method, url string, b io.Reader) (*http.Response, error) {

	req, err := http.NewRequest(method, url, b)
	if err != nil {
		h.l.Error("error creating request", zap.Error(err))
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+h.accessToken)

	resp, err := h.client.Do(req)
	if err != nil {
		h.l.Error("error sending request", zap.Error(err))
		return nil, err
	}
	return resp, nil
}
