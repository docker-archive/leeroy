package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
	"github.com/docker/leeroy/github"
)

// Commit describes information in a commit
type Commit struct {
	CommentsURL string `json:"comments_url,omitempty"`
	HTMLURL     string `json:"html_url,omitempty"`
	Sha         string `json:"sha,omitempty"`
	URL         string `json:"url,omitempty"`
}

func (c Config) getBuilds(baseRepo string, isCustom bool) (builds []Build, err error) {
	for _, build := range c.Builds {
		if build.Repo == baseRepo && isCustom == build.Custom {
			builds = append(builds, build)
		}
	}

	if len(builds) <= 0 {
		return builds, fmt.Errorf("Could not find config for %s", baseRepo)
	}

	return builds, nil
}

func (c Config) getBuildByJob(job string) (build Build, err error) {
	for _, build := range c.Builds {
		if build.Job == job {
			return build, nil
		}
	}

	return build, fmt.Errorf("Could not find config for %s", job)
}

func (c Config) getBuildByContextAndRepo(context, repo string) (build Build, err error) {
	if context == "" {
		context = DEFAULTCONTEXT
	}

	for _, build := range c.Builds {
		if build.Context == context && build.Repo == repo {
			return build, nil
		}
	}

	return build, fmt.Errorf("Could not find config for context: %s, repo: %s", context, repo)
}

func (c Config) addGithubComment(repoName, pr, comment string) error {
	// parse git repo for username
	// and repo name
	r := strings.SplitN(repoName, "/", 2)
	if len(r) < 2 {
		return fmt.Errorf("repo name could not be parsed: %s", repoName)
	}

	// initialize github client
	gh := octokat.NewClient()
	gh = gh.WithToken(c.GHToken)
	repo := octokat.Repo{
		Name:     r[1],
		UserName: r[0],
	}

	// add comment to the PR
	if _, err := gh.AddComment(repo, pr, comment); err != nil {
		return fmt.Errorf("adding comment to %s#%s failed: %v", repoName, pr, err)
	}

	return nil
}

func (c Config) updateGithubStatus(repoName, context, sha, state, desc, buildURL string) error {
	// parse git repo for username
	// and repo name
	r := strings.SplitN(repoName, "/", 2)
	if len(r) < 2 {
		return fmt.Errorf("repo name could not be parsed: %s", repoName)
	}

	// initialize github client
	gh := octokat.NewClient()
	gh = gh.WithToken(c.GHToken)
	repo := octokat.Repo{
		Name:     r[1],
		UserName: r[0],
	}

	status := &octokat.StatusOptions{
		State:       state,
		Description: desc,
		URL:         buildURL,
		Context:     context,
	}
	if _, err := gh.SetStatus(repo, sha, status); err != nil {
		return fmt.Errorf("setting status for repo: %s, sha: %s failed: %v", repoName, sha, err)
	}

	logrus.Infof("Setting status on %s %s to %s for %s succeeded", repoName, sha, state, context)
	return nil
}

func hasStatus(gh *octokat.Client, repo octokat.Repo, sha, context string) bool {
	statuses, err := gh.Statuses(repo, sha, &octokat.Options{})
	if err != nil {
		logrus.Warnf("getting status for %s for %s/%s failed: %v", sha, repo.UserName, repo.Name, err)
		return false
	}

	for _, status := range statuses {
		if status.Context == context && status.State == "success" {
			return true
		}
	}

	return false
}

func (c Config) getShas(owner, name, context string, number int) (shas []string, pr *octokat.PullRequest, err error) {
	// initialize github client
	gh := octokat.NewClient()
	gh = gh.WithToken(c.GHToken)
	repo := octokat.Repo{
		Name:     name,
		UserName: owner,
	}

	// get the pull request so we can get the commits
	pr, err = gh.PullRequest(repo, strconv.Itoa(number), &octokat.Options{})
	if err != nil {
		return shas, pr, fmt.Errorf("getting pull request %d for %s/%s failed: %v", number, owner, name, err)
	}

	// check which commits we want to get
	// from the original flag --build-commits
	if c.BuildCommits == "all" || c.BuildCommits == "new" {

		// get the commits url
		req, err := http.Get(pr.CommitsURL)
		if err != nil {
			return shas, pr, err
		}
		defer req.Body.Close()

		// parse the response
		var commits []Commit
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&commits); err != nil {
			return shas, pr, fmt.Errorf("parsing the response from %s failed: %v", pr.CommitsURL, err)
		}

		// append the commit shas
		for _, commit := range commits {
			// if we only want the new shas
			// check to make sure the status
			// has not been set before appending
			if c.BuildCommits == "new" {
				if hasStatus(gh, repo, commit.Sha, context) {
					continue
				}
			}

			shas = append(shas, commit.Sha)
		}
	} else {
		// this is the case where buildCommits == "last"
		// just get the sha of the pr
		shas = append(shas, pr.Head.Sha)
	}

	return shas, pr, nil
}

func (c Config) scheduleJenkinsBuild(baseRepo string, number int, build Build) error {
	// setup the jenkins client
	j := &config.Jenkins

	// make sure we even want to build
	if build.Job == "" {
		return nil
	}

	// cancel any existing builds if we can, before sheduling another
	/*
		if err := j.CancelBuildsForPR(build.Job, strconv.Itoa(number)); err != nil {
			logrus.Warnf("Trying to cancel existing builds for job %s, pr %d failed: %v", build.Job, number, err)
		}
	*/

	// find the comments about failed builds and remove them
	if err := config.removeFailedBuildComment(baseRepo, build.Job, number); err != nil {
		logrus.Error(err)
	}

	// parse git repo for username
	// and repo name
	r := strings.SplitN(baseRepo, "/", 2)
	if len(r) < 2 {
		return fmt.Errorf("repo name could not be parsed: %s", baseRepo)
	}

	// get the shas to build
	shas, pr, err := c.getShas(r[0], r[1], build.Context, number)
	if err != nil {
		return err
	}

	for _, sha := range shas {
		// update the github status
		if err := c.updateGithubStatus(baseRepo, build.Context, sha, "pending", "Jenkins build is being scheduled", c.Jenkins.Baseurl); err != nil {
			return err
		}

		// setup the parameters
		htmlURL := fmt.Sprintf("https://github.com/%s/pull/%d", baseRepo, pr.Number)
		headRepo := fmt.Sprintf("%s/%s", pr.Head.Repo.Owner.Login, pr.Head.Repo.Name)
		parameters := fmt.Sprintf("GIT_BASE_REPO=%s&GIT_HEAD_REPO=%s&GIT_SHA1=%s&GITHUB_URL=%s&PR=%d&BASE_BRANCH=%s", baseRepo, headRepo, sha, htmlURL, pr.Number, pr.Base.Ref)
		// schedule the build
		if err := j.BuildWithParameters(build.Job, parameters); err != nil {
			return fmt.Errorf("scheduling jenkins build failed: %v", err)
		}
	}

	return nil
}

func (c Config) getFailedPRs(context, repoName string) (nums []int, err error) {
	// parse git repo for username
	// and repo name
	r := strings.SplitN(repoName, "/", 2)
	if len(r) < 2 {
		return nums, fmt.Errorf("repo name could not be parsed: %s", repoName)
	}

	// initialize github client
	gh := octokat.NewClient()
	gh = gh.WithToken(c.GHToken)
	repo := octokat.Repo{
		Name:     r[1],
		UserName: r[0],
	}

	// get pull requests
	prs, err := gh.PullRequests(repo, &octokat.Options{
		Params: map[string]string{
			"state":    "open",
			"per_page": "100",
		},
	})
	if err != nil {
		return nums, fmt.Errorf("requesting open repos for %s failed: %v", repoName, err)
	}

	for _, pr := range prs {
		if !hasStatus(gh, repo, pr.Head.Sha, context) {
			nums = append(nums, pr.Number)
		}
	}

	return nums, nil
}

func (c Config) removeFailedBuildComment(repoName, job string, pr int) error {
	// parse git repo for username
	// and repo name
	r := strings.SplitN(repoName, "/", 2)
	if len(r) < 2 {
		return fmt.Errorf("repo name could not be parsed: %s", repoName)
	}

	// initialize github client
	g := github.GitHub{
		AuthToken: c.GHToken,
		User:      c.GHUser,
	}
	repo := octokat.Repo{
		Name:     r[1],
		UserName: r[0],
	}

	content, err := g.GetContent(repo, pr, true)
	if err != nil {
		return fmt.Errorf("getting pull request content failed: %v", err)
	}

	// find the comments about failed builds and remove them
	if comment := content.FindComment(fmt.Sprintf("Job: %s [FAILED", job), c.GHUser); comment != nil {
		if err := g.Client().RemoveComment(repo, comment.Id); err != nil {
			return fmt.Errorf("removing comment from %s#%d for %s failed: %v", repoName, pr, job, err)
		}
	}

	return nil
}
