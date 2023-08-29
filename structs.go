package sfimport

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Session struct {
	sync.RWMutex
	clientKey     string
	clientSecret  string
	username      string
	password      string
	securityToken string
	instanceURL   string
	authURL       string
	tokenURL      string
	sobjectsURL   string
	Client        *http.Client
	wsConn        *websocket.Conn
	wsMutex       sync.Mutex
	Dialer        *websocket.Dialer
	UserAgent     string
	access_token  string
	refresh_token string
}

type FieldMetadata struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type MetadataResponse struct {
	Fields []FieldMetadata `json:"fields"`
}

type FieldAPILabelMapping map[string]string
