package jenkins

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

func (c *Client) ScheduleBuild() {
}
