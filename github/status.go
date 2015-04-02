package github

import "github.com/crosbymichael/octokat"

func (g GitHub) successStatus(repo octokat.Repo, sha, context, description string) error {
	_, err := g.Client().SetStatus(repo, sha, &octokat.StatusOptions{
		State:       "success",
		Context:     context,
		Description: description,
	})
	return err
}

func (g GitHub) failureStatus(repo octokat.Repo, sha, context, description, targetUrl string) error {
	_, err := g.Client().SetStatus(repo, sha, &octokat.StatusOptions{
		State:       "failure",
		Context:     context,
		Description: description,
		URL:         targetUrl,
	})
	return err
}
