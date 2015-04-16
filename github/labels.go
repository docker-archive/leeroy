package github

import (
	"strings"

	"github.com/crosbymichael/octokat"
)

func (g GitHub) toggleLabels(repo octokat.Repo, issueNum int, labelToRemove, labelToAdd string) error {
	exist, err := g.labelExist(repo, issueNum, labelToAdd)
	if err != nil {
		return err
	}

	if !exist {
		if err := g.removeLabel(repo, issueNum, labelToRemove); err != nil {
			return err
		}
		if err := g.addLabel(repo, issueNum, labelToAdd); err != nil {
			return err
		}
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
		if err := g.Client().RemoveLabel(repo, &issue, label); err != nil && !strings.Contains(err.Error(), "Label does not exist") {
			return err
		}
	}

	return nil
}

func (g GitHub) labelExist(repo octokat.Repo, issueNum int, label string) (bool, error) {
	i, err := g.Client().Issue(repo, issueNum, &octokat.Options{})
	if err != nil {
		return false, err
	}

	if i.Labels == nil || len(i.Labels) == 0 {
		return false, nil
	}

	for _, l := range i.Labels {
		if l.Name == label {
			return true, nil
		}
	}

	return false, nil
}
