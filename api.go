package sfimport

import (
	"log"
	"net/http"
)

func (s *Session) InitiateConnection1() error {

	//var dialer *websocket.Dialer
	//var mx sync.Mutex
	//var sessionID string

	req, err := http.NewRequest("GET", s.endpoint, nil)
	if err != nil {
		log.Fatalln("Error creating request:", err)
		return err
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Fatalln("Error making request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalln("Request failed with status:", resp.Status)
		//return nil
	}

	return nil
}
