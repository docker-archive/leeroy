package github

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
)

// IsMergeable makes sure the pull request can be merged
func (g GitHub) IsMergeable(pr *PullRequest, failureLink string) (mergeable bool, err error) {
	// assume the PR is mergable unless we specifically set to false
	// because mergable true is equivalent to skip
	mergeable = true

	// we only want the prs that are opened/synchronized
	if !pr.Hook.IsOpened() && !pr.Hook.IsSynchronize() {
		return mergeable, nil
	}

	jobName := fmt.Sprintf("%s/is-mergable", pr.Repo.UserName)
	commentType := "merge conflicts"
	if !isMergeable(pr) {
		mergeable = false
		logrus.Debugf("Found pr %d was not mergable, going to add comment", pr.Hook.Number)

		// add a comment
		comment := "Looks like we would not be able to merge this PR because of merge conflicts. Please rebase, fix conflicts, and force push to your branch."
		if err := g.addUniqueComment(pr.Repo, strconv.Itoa(pr.Hook.Number), comment, commentType, pr.Content); err != nil {
			return mergeable, err
		}

		// set the status
		if err := g.failureStatus(pr.Repo, pr.Head.Sha, jobName, "This PR is not mergable, please fix conflicts.", failureLink); err != nil {
			return mergeable, err
		}

		return mergeable, nil
	}

	// otherwise try to find the comment and remove it
	if err := g.removeComment(pr.Repo, commentType, pr.Content); err != nil {
		return mergeable, err
	}

	return mergeable, nil
}

func isMergeable(pr *PullRequest) bool {
	// this is kinda hacky because we made Mergeable a *bool
	if pr.Mergeable != nil && *pr.Mergeable == false {
		return false
	}

	return true
}
