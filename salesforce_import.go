package sfimport

import (
	"net/http"
	"time"
)

const VERSION = "0.0.1"

func CreateSession(clientKey, clientSecret, username, password, securityToken, instanceURL string) (sesh *Session, err error) {
	sesh = &Session{
		clientKey:     clientKey,
		clientSecret:  clientSecret,
		username:      username,
		password:      password,
		securityToken: securityToken,
		instanceURL:   instanceURL,
		authURL:       instanceURL + "/services/oauth2/authorize",
		tokenURL:      instanceURL + "/services/oauth2/token",
		UserAgent:     "SF-Import (https://github.com/AmitSuresh/SF-Import, v" + VERSION + ")",
		Client:        &http.Client{Timeout: (20 * time.Second)},
	}
	return
}
