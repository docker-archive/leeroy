package github

import "github.com/crosbymichael/octokat"

func (g GitHub) toggleLabels(repo octokat.Repo, issueNum int, labelToRemove, labelToAdd string) error {
	if err := g.removeLabel(repo, issueNum, labelToRemove); err != nil {
		return err
	}
	if err := g.addLabel(repo, issueNum, labelToAdd); err != nil {
		return err
	}
	return nil
}

func (g GitHub) addLabel(repo octokat.Repo, issueNum int, labels ...string) error {
	issue := octokat.Issue{
		Number: issueNum,
	}

	return g.Client().ApplyLabel(repo, &issue, labels)
}

func (g GitHub) removeLabel(repo octokat.Repo, issueNum int, labels ...string) error {
	issue := octokat.Issue{
		Number: issueNum,
	}

	for _, label := range labels {
		return g.Client().RemoveLabel(repo, &issue, label)
	}

	return nil
}
