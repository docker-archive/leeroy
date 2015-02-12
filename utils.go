package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
	"github.com/jfrazelle/leeroy/jenkins"
)

type Commit struct {
	CommentsURL string `json:"comments_url,omitempty"`
	HtmlURL     string `json:"html_url,omitempty"`
	Sha         string `json:"sha,omitempty"`
	URL         string `json:"url,omitempty"`
}

func getBuild(baseRepo string, builds []Build) (build Build, err error) {
	for _, build := range builds {
		if build.Repo == baseRepo {
			return build, nil
		}
	}

	return build, fmt.Errorf("Could not find config for %s", baseRepo)
}

func updateGithubStatus(repoName, context, sha, state, desc, buildUrl string) error {
	// parse git repo for username
	// and repo name
	r := strings.SplitN(repoName, "/", 2)
	if len(r) < 2 {
		return fmt.Errorf("repo name could not be parsed: %s", repoName)
	}

	// initialize github client
	gh := octokat.NewClient()
	gh = gh.WithToken(ghtoken)
	repo := octokat.Repo{
		Name:     r[1],
		UserName: r[0],
	}

	status := &octokat.StatusOptions{
		State:       state,
		Description: desc,
		URL:         buildUrl + "console",
		Context:     context,
	}
	if _, err := gh.SetStatus(repo, sha, status); err != nil {
		return fmt.Errorf("setting status for repo: %s, sha: %s failed: %v", repoName, sha, err)
	}

	log.Infof("Setting status on %s %s to %s succeeded", repoName, sha, state)
	return nil
}

func hasStatus(gh *octokat.Client, repo octokat.Repo, sha string) bool {
	statuses, err := gh.Statuses(repo, sha, &octokat.Options{})
	if err != nil {
		log.Warnf("getting status for %s for %s/%s failed: %v", sha, repo.UserName, repo.Name, err)
		return false
	}

	for _, status := range statuses {
		if status.Context == Context && status.State == "success" {
			return true
		}
	}

	return false
}

func getShas(owner, name string, number int) (shas []string, pr *octokat.PullRequest, err error) {
	// initialize github client
	gh := octokat.NewClient()
	gh = gh.WithToken(ghtoken)
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
	if buildCommits == "all" || buildCommits == "new" {

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
			if buildCommits == "new" {
				if hasStatus(gh, repo, commit.Sha) {
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

func scheduleJenkinsBuild(baseRepo string, number int) error {
	// parse git repo for username
	// and repo name
	r := strings.SplitN(baseRepo, "/", 2)
	if len(r) < 2 {
		return fmt.Errorf("repo name could not be parsed: %s", baseRepo)
	}

	// get the shas to build
	shas, pr, err := getShas(r[0], r[1], number)
	if err != nil {
		return err
	}

	for _, sha := range shas {
		// get the build
		build, err := getBuild(baseRepo, builds)
		if err != nil {
			return err
		}

		// update the github status
		if err := updateGithubStatus(baseRepo, build.Context, sha, "pending", "Jenkins build is being scheduled", ""); err != nil {
			return err
		}

		// setup the jenkins client
		j := jenkins.New(jenkinsUri, jenkinsUser, jenkinsPass)
		// setup the parameters
		htmlUrl := fmt.Sprintf("https://github.com/%s/pull/%d")
		headRepo := fmt.Sprintf("%s/%s", pr.Head.Repo.Owner.Login, pr.Head.Repo.Name)
		parameters := fmt.Sprintf("GIT_BASE_REPO=%s&GIT_HEAD_REPO=%s&GIT_SHA1=%s&GITHUB_URL=%s&PR=%d&BASE_BRANCH=%s", baseRepo, headRepo, sha, htmlUrl, pr.Number, pr.Base.Ref)
		// schedule the build
		if err := j.BuildWithParameters(build.Job, parameters); err != nil {
			return fmt.Errorf("scheduling jenkins build failed: %v", err)
		}
	}

	return nil
}
