package github

import (
	"strings"

	"github.com/crosbymichael/octokat"
)

type labels map[string]bool

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

func (g GitHub) addLabel(repo octokat.Repo, issueNum int, labelsToAdd ...string) error {
	issue, issueLabels, err := g.issueWithLabels(repo, issueNum)
	if err != nil {
		return err
	}

	var l []string
	for _, label := range labelsToAdd {
		if !issueLabels[label] {
			l = append(l, label)
		}
	}

	return g.Client().ApplyLabel(repo, issue, l)
}

func (g GitHub) removeLabel(repo octokat.Repo, issueNum int, labelsToRemove ...string) error {
	issue, issueLabels, err := g.issueWithLabels(repo, issueNum)
	if err != nil {
		return err
	}

	for _, label := range labelsToRemove {
		if issueLabels[label] {
			if err := g.Client().RemoveLabel(repo, issue, label); err != nil && !strings.Contains(err.Error(), "Label does not exist") {
				return err
			}
		}
	}

	return nil
}

func (g GitHub) issueWithLabels(repo octokat.Repo, issueNum int) (*octokat.Issue, labels, error) {
	issue, err := g.Client().Issue(repo, issueNum, &octokat.Options{})
	if err != nil {
		return nil, labels{}, err
	}

	l := labels{}
	for _, label := range issue.Labels {
		l[label.Name] = true
	}

	return issue, l, nil
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
