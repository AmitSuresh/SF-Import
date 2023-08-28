package sfimport

func CreateSession(username, password, securityToken, instanceURL string) (s *Session, err error) {
	sesh := &Session{}
	sesh.username = username
	sesh.password = password
	sesh.securityToken = securityToken
	sesh.instanceURL = instanceURL
	sesh.endpoint = instanceURL + "/services/data/v53.0/jobs/ingest"

	return
}
