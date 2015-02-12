package jenkins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	Baseurl  string
	username string
	password string
}

type JenkinsResponse struct {
	Name  string       `json:"name"`
	Build JenkinsBuild `json:"build"`
}

type JenkinsBuild struct {
	Number     int                    `json:"number"`
	Url        string                 `json:"full_url"`
	Phase      string                 `json:"phase"`
	Status     string                 `json:"status"`
	Parameters JenkinsBuildParameters `json:"parameters"`
}

type JenkinsBuildParameters struct {
	GitBaseRepo string `json:"GIT_BASE_REPO"`
	GitSha      string `json:"GIT_SHA1"`
}

type Request struct {
	Parameters []map[string]string `json:"parameter"`
}

// Sets the authentication for the Jenkins client
// Password can be an API token as described in:
// https://wiki.jenkins-ci.org/display/JENKINS/Authenticating+scripted+clients
func New(uri, username, password string) *Client {
	return &Client{
		Baseurl:  uri,
		username: username,
		password: password,
	}
}

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
	req.SetBasicAuth(c.username, c.password)

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

func (c *Client) BuildWithParameters(job string, parameters string) error {
	// set up the request
	url := fmt.Sprintf("%s/job/%s/buildWithParameters?%s", c.Baseurl, job, parameters)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}

	// add the auth
	req.SetBasicAuth(c.username, c.password)

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
