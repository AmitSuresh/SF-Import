package sfimport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (sesh *Session) InitiateConnection() error {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", sesh.username)
	data.Set("password", sesh.password+sesh.securityToken)
	data.Set("client_id", sesh.clientKey)
	data.Set("client_secret", sesh.clientSecret)

	req, err := http.NewRequest("POST", sesh.tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", sesh.UserAgent) // Set your user agent
	req.Header.Set("Accept", "*/*")

	sesh.Client = &http.Client{}
	resp, err := sesh.Client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status:", resp.Status)
		return fmt.Errorf("Request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}

	// Assuming the response body contains JSON data, you can unmarshal it
	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return err
	}

	// Now you can access the response data as needed, e.g., access_token, instance_url, etc.
	fmt.Println("id:", responseData["id"])
	fmt.Println("Access Token:", responseData["access_token"])
	fmt.Println("Instance URL:", responseData["instance_url"])

	return nil
}
