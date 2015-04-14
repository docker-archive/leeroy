package github

import (
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

func (g GitHub) IsMergeable(prHook *octokat.PullRequestHook) (mergeable bool, err error) {
	// assume the PR is mergable unless we specifically set to false
	// because mergable true is equivalent to skip
	mergeable = true

	// we only want the prs that are opened/synchronized
	if !prHook.IsOpened() && !prHook.IsSynchronize() {
		return mergeable, nil
	}

	// get the PR
	pr := prHook.PullRequest
	repo := getRepo(prHook.Repo)

	content, err := g.getPullRequestContent(repo, prHook.Number)
	if err != nil {
		return mergeable, err
	}

	commentType := "merge conflicts"
	if !isMergeable(pr) {
		mergeable = false
		log.Debugf("Found pr %d was not mergable, going to add comment", prHook.Number)

		// add a comment
		comment := "Looks like we would not be able to merge this PR because of merge conflicts. Please rebase, fix conflicts, and force push to your branch."
		if err := g.addUniqueComment(repo, strconv.Itoa(prHook.Number), comment, commentType, content); err != nil {
			return mergeable, err
		}

		// set the status
		if err := g.failureStatus(repo, pr.Head.Sha, "docker/is-mergable", "This PR is not mergable, please fix conflicts.", "https://docs.docker.com/project/work-issue/"); err != nil {
			return mergeable, err
		}

		return mergeable, nil
	}

	// otherwise try to find the comment and remove it
	if err := g.removeComment(repo, pr, commentType, content); err != nil {
		return mergeable, err
	}

	return mergeable, nil
}

func isMergeable(pr *octokat.PullRequest) bool {
	// this is kinda hacky because we made Mergeable a *bool
	if pr.Mergeable != nil && *pr.Mergeable == false {
		return false
	}

	return true
}
