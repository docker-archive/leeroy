package github

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

func (g GitHub) DcoVerified(prHook *octokat.PullRequestHook) (bool, error) {
	// we only want the prs that are opened
	if !prHook.IsOpened() {
		return false, nil
	}

	// get the PR
	pr := prHook.PullRequest
	repo := getRepo(prHook.Repo)

	content, err := g.getPullRequestContent(repo, prHook.Number)
	if err != nil {
		return false, err
	}

	// we only want apply labels
	// to opened pull requests
	var labels []string

	//check if it's a proposal
	isProposal := strings.Contains(strings.ToLower(pr.Title), "proposal")
	switch {
	case isProposal:
		labels = []string{"status/1-needs-design-review"}
	case content.IsDocsOnly():
		labels = []string{"status/3-needs-docs-review"}
	default:
		labels = []string{"status/0-needs-triage"}
	}

	//add labels if there are any
	if len(labels) > 0 {
		log.Debugf("Adding labels %#v to pr %d", labels, prHook.Number)

		if err := g.addLabel(repo, prHook.Number, labels...); err != nil {
			return false, err
		}

		log.Infof("Added labels %#v to pr %d", labels, prHook.Number)
	}

	var verified bool
	if content.CommitsSigned() {
		if err := g.toggleLabels(repo, prHook.Number, "dco/no", "dco/yes"); err != nil {
			return false, err
		}

		if err := g.successStatus(repo, pr.Head.Sha, "docker-dco", "All commits signed"); err != nil {
			return false, err
		}

		verified = true
	} else {
		if err := g.toggleLabels(repo, prHook.Number, "dco/yes", "dco/no"); err != nil {
			return false, err
		}

		if err := g.failureStatus(repo, pr.Head.Sha, "docker-dco", "Some commits without signature", "https://github.com/docker/docker/blob/master/CONTRIBUTING.md#sign-your-work"); err != nil {
			return false, err
		}
	}

	return verified, nil
}

func getRepo(repo *octokat.Repository) octokat.Repo {
	return octokat.Repo{
		Name:     repo.Name,
		UserName: repo.Owner.Login,
	}
}