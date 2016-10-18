package jenkins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
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
	PR          string `json:"PR"`
}

// Request describes a request to jenkins
type Request struct {
	Parameters []map[string]string `json:"parameter,omitempty"`
}

// JobBuildsResponse describes a response for a job's builds.
type JobBuildsResponse struct {
	Builds []RecentBuild `json:"builds,omitempty"`
}

// RecentBuild describes a build from the Jenkins API.
type RecentBuild struct {
	ID        string    `json:"id,omitempty"`
	Actions   []Action  `json:"actions,omitempty"`
	Building  bool      `json:"building,omitempty"`
	Timestamp time.Time `json:"timstamp,omitempty"`
	NodeName  string    `json:"builtOn,omitempty"`
}

// Action defines the action for a build.
type Action struct {
	Parameters []Parameter `json:"parameters,omitempty"`
}

// Parameter defines the parameters for a build.
type Parameter struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// QueuedBuildsResponse describes a response for a job's builds.
type QueuedBuildsResponse struct {
	Builds []QueuedBuild `json:"items,omitempty"`
}

// QueuedBuild represents a build in the queue.
type QueuedBuild struct {
	ID      int       `json:"id,omitempty"`
	Actions []Action  `json:"actions,omitempty"`
	Task    QueueTask `json:"task,omitempty"`
}

// QueueTask is a task associated with a build in the queue.
type QueueTask struct {
	Name string `json:"name,omitempty"`
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

// BuildPipeline is just BuildWithParameters but for a Pipeline job instead.
func (c *Client) BuildPipeline(job string, prNumber int, prRef string) error {
	subJobName := prRef
	if prNumber != 0 {
		subJobName = fmt.Sprintf("PR-%d", prNumber)
	}
	url := fmt.Sprintf("%s/job/%s/job/%s/build", c.Baseurl, job, subJobName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.Username, c.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("jenkins post to %s responded with status %d", url, resp.StatusCode)
	}

	return nil
}

// CancelBuildsForPR cancels any queued or running builds for a PR.
func (c *Client) CancelBuildsForPR(job, pr string) error {
	if env := os.Getenv("LEEROY_KEEP_OLD_BUILD_RUNNING"); env != "" {
		return errors.New("LEEROY_KEEP_OLD_BUILD_RUNNING is set")
	}

	var e string

	// first check the queue
	q, err := c.GetQueuedBuildForPR(job, pr)
	if err != nil {
		e = fmt.Sprintf("Getting queued build for job %s, pr %s failed: %v; ", job, pr, err)
	} else if q != nil {
		// if it is not nil then we found a matching build, cancel it
		if err := c.CancelBuild(job, strconv.Itoa(q.ID), true); err != nil {
			e = fmt.Sprintf("cancelling queued build for job %s, pr %s failed: %v; ", job, pr, err)
		}
		logrus.Infof("Cancelled queued build (%d) for job %s, pr %s", q.ID, job, pr)
	}

	// check running builds
	b, err := c.GetRunningBuildForPR(job, pr)
	if err != nil {
		e += fmt.Sprintf("Getting running build for job %s, pr %s failed: %v;", job, pr, err)
	} else if b != nil {
		// if it is not nil then we found a matching build, cancel it
		if err := c.CancelBuild(job, b.ID, false); err != nil {
			e += fmt.Sprintf("cancelling running build for job %s, pr %s failed: %v;", job, pr, err)
		}
		logrus.Infof("Cancelled running build (%s) for job %s, pr %s", b.ID, job, pr)
	}

	if e != "" {
		return errors.New(e)
	}

	return nil
}

// CancelBuild cancels/stops a running or queued build.
func (c *Client) CancelBuild(job, id string, isQueued bool) error {
	// set up the request
	url := fmt.Sprintf("%s/job/%s/%s/stop", c.Baseurl, job, id)
	if isQueued {
		url = fmt.Sprintf("%s/queue/cancelItem?id=%s", c.Baseurl, id)
	}
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
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return fmt.Errorf("jenkins post to %s responded with status %d", url, resp.StatusCode)
	}

	return nil
}

// GetRunningBuildForPR returns the running build for a Jenkins job and PR if there is one.
func (c *Client) GetRunningBuildForPR(job, pr string) (*RecentBuild, error) {
	builds, err := c.GetBuilds(job)
	if err != nil {
		return nil, err
	}

	for _, build := range builds {
		if build.Building {
			for _, a := range build.Actions {
				for _, p := range a.Parameters {
					if p.Name == "PR" && p.Value == pr {
						return &build, nil
					}
				}
			}
		}
	}

	return nil, nil
}

// GetBuilds gets the builds for a Jenkins job.
func (c *Client) GetBuilds(job string) (b []RecentBuild, err error) {
	// set up the request
	url := fmt.Sprintf("%s/job/%s/api/json?tree=%s", c.Baseurl, job, url.QueryEscape("builds[builtOn,actions[parameters[name,value]],timestamp,id,building]"))
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return b, err
	}

	// add the auth
	req.SetBasicAuth(c.Username, c.Token)

	// do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return b, err
	}
	defer resp.Body.Close()

	// check the status code
	// it should be 200
	if resp.StatusCode != 200 {
		return b, fmt.Errorf("jenkins get builds for %s request to %s responded with status %d", job, url, resp.StatusCode)
	}

	var r JobBuildsResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return b, fmt.Errorf("decoding json response from builds from %s failed: %v", url, err)
	}

	return r.Builds, nil
}

// GetQueuedBuildForPR returns the queued build for a Jenkins job and PR if there is one.
func (c *Client) GetQueuedBuildForPR(job, pr string) (*QueuedBuild, error) {
	// set up the request
	url := fmt.Sprintf("%s/queue/api/json?tree=%s", c.Baseurl, url.QueryEscape("items[id,task[name]]"))
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}

	// add the auth
	req.SetBasicAuth(c.Username, c.Token)

	// do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// check the status code
	// it should be 200
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("jenkins get queued builds request to %s responded with status %d", url, resp.StatusCode)
	}

	var r QueuedBuildsResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decoding json response from queued builds from %s failed: %v", url, err)
	}

	// loop through and collect only the ones that task name matches the job.
	for _, build := range r.Builds {
		if build.Task.Name == job {
			for _, a := range build.Actions {
				for _, p := range a.Parameters {
					if p.Name == "PR" && p.Value == pr {
						return &build, nil
					}
				}
			}
		}
	}

	return nil, nil
}

// GetBuildLog returns the consoleText for a Jenkins build.
func (c *Client) GetBuildLog(job string, id int) (string, error) {
	// set up the request
	url := fmt.Sprintf("%s/job/%s/%d/consoleText", c.Baseurl, job, id)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// check the status code
	// it should be 200
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("jenkins get logs request to %s responded with status %d", url, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading body from logs response to %s failed: %v", url, err)
	}

	return parseFailedBuildLog(job, url, string(body)), nil
}

func parseFailedBuildLog(job, url, log string) string {
	testComment := fmt.Sprintf(`Job: %s [FAILED](%s):

~~~console

`, job, strings.Replace(url, "consoleText", "console", 1))
	// set the chars to resturn around the index where the line was matched
	sl := 500

	// first try to find FAIL in the log
	re := regexp.MustCompile(`FAIL((.)*?)(\n|\r)`)
	m := re.FindAllStringIndex(log, 1)

	if len(m) == 0 {
		// try another way, by looking for the end before the PostBuildScript
		re = regexp.MustCompile(`PostBuildScript((.)*?)(\n|\r)`)
		m = re.FindAllStringIndex(log, 1)
	}

	if len(m) == 0 {
		// if still empty just return empty comment
		return ""
	}

	// get the post build script index
	pbIndex := strings.Index(log, "Now starting POST-BUILD steps")

	for _, index := range m {
		// get a few lines around the index
		start := 0
		if (index[0] - sl) > start {
			start = index[0] - sl
		}
		end := len(log) - 1
		if pbIndex > 5 {
			// set the end to be that instead, since we don't need to include
			// the post build script logs in the comment
			end = pbIndex - 5
		}
		if index[1]+sl < end {
			end = index[1] + sl
		}
		testComment += fmt.Sprintf("---\n%s\n---\n\n", log[start:end])
	}

	return testComment + "~~~"
}
