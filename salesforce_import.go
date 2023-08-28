package sfimport

func CreateSession(username, password, securityToken, instanceURL string) (s *Session, err error) {
	sesh := &Session{}
	s.username = username
	s.password = password
	s.securityToken = securityToken
	s.instanceURL = instanceURL
	s.endpoint = instanceURL + "/services/data/v53.0/jobs/ingest"

	return
}
