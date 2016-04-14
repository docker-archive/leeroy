package octokat

import (
	"fmt"
)

type Contributor struct {
	Total int `json:"total,omitempty"`
	Weeks []struct {
		WeekStart int `json:"w,omitempty"`
		Additions int `json:"a,omitempty"`
		Deletions int `json:"d,omitempty"`
		Commits   int `json:"c,omitempty"`
	} `json:"weeks,omitempty"`
	Author struct {
		Login             string `json:"login,omitempty"`
		Id                int    `json:"id,omitempty"`
		AvatarURL         string `json:"avatar_url,omitempty"`
		GravatarID        string `json:"gravatar_id,omitempty"`
		URL               string `json:"url,omitempty"`
		HTMLURL           string `json:"html_url,omitempty"`
		FollowersURL      string `json:"followers_url,omitempty"`
		FollowingURL      string `json:"following_url,omitempty"`
		GistsURL          string `json:"gists_url,omitempty"`
		StarredURL        string `json:"starred_url,omitempty"`
		SubscriptionsURL  string `json:"subscriptions_url,omitempty"`
		OrganizationsURL  string `json:"organizations_url,omitempty"`
		ReposURL          string `json:"repos_url,omitempty"`
		EventsURL         string `json:"events_url,omitempty"`
		ReceivedEventsURL string `json:"received_events_url,omitempty"`
		Type              string `json:"type,omitempty"`
		SiteAdmin         bool   `json:"site_admin,omitempty"`
	} `json:"author,omitempty"`
}

// Get a list of contributors
//
//http://developer.github.com/v3/repos/statistics/#contributors
func (c *Client) Contributors(repo Repo, options *Options) (contributors []*Contributor, err error) {
	path := fmt.Sprintf("repos/%s/stats/contributors", repo)
	err = c.jsonGet(path, options, &contributors)
	return
}
