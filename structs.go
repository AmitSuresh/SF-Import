package sfimport

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Session struct {
	sync.RWMutex
	username      string
	password      string
	securityToken string
	instanceURL   string
	endpoint      string
	Client        *http.Client
	wsConn        *websocket.Conn
	wsMutex       sync.Mutex
	Dialer        *websocket.Dialer
}
