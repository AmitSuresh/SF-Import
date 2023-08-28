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
	Client        *http.Client
	wsConn        *websocket.Conn
	wsMutex       sync.Mutex
	Dialer        *websocket.Dialer
	UserAgent     string
}
