package jenkins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client contains the information for connecting to a jenkins instance
type Client struct {
	Baseurl  string `json:"base_url"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// Response describes the response returned by jenkins
type Response struct {
	Name  string `json:"name"`
	Build Build  `json:"build"`
}

// Build describes a jenkins build
type Build struct {
	Number     int             `json:"number"`
	URL        string          `json:"full_url"`
	Phase      string          `json:"phase"`
	Status     string          `json:"status"`
	Parameters BuildParameters `json:"parameters"`
}

// BuildParameters decribes the jenkins build parameters
type BuildParameters struct {
	GitBaseRepo string `json:"GIT_BASE_REPO"`
	GitSha      string `json:"GIT_SHA1"`
}

// Request describes a request to jenkins
type Request struct {
	Parameters []map[string]string `json:"parameter"`
}

// New sets the authentication for the Jenkins client
// Password can be an API token as described in:
// https://wiki.jenkins-ci.org/display/JENKINS/Authenticating+scripted+clients
func New(uri, username, token string) *Client {
	return &Client{
		Baseurl:  uri,
		Username: username,
		Token:    token,
	}
}

// Build sends a build request to jenkins
func (c *Client) Build(job string, data Request) error {
	// encode the request data
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// set up the request
	url := fmt.Sprintf("%s/job/%s/build", c.Baseurl, job)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(d))
	if err != nil {
		return err
	}

	// add the auth
	req.SetBasicAuth(c.Username, c.Token)

	// do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// check the status code
	// it should be 201
	if resp.StatusCode != 201 {
		return fmt.Errorf("jenkins post to %s responded with status %d, data: %s", url, resp.StatusCode, string(d))
	}

	return nil
}

// BuildWithParameters sends a build request with parameters to jenkins
func (c *Client) BuildWithParameters(job string, parameters string) error {
	// set up the request
	url := fmt.Sprintf("%s/job/%s/buildWithParameters?%s", c.Baseurl, job, parameters)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}

	// add the auth
	req.SetBasicAuth(c.Username, c.Token)

	// do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// check the status code
	// it should be 201
	if resp.StatusCode != 201 {
		return fmt.Errorf("jenkins post to %s responded with status %d", url, resp.StatusCode)
	}

	return nil
}
