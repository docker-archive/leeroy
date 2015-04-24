package github

import (
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

func (g GitHub) addDCOUnsignedComment(repo octokat.Repo, pr *octokat.PullRequest, content *pullRequestContent) error {
	comment := `Can you please sign your commits following these rules:
https://github.com/docker/docker/blob/master/CONTRIBUTING.md#sign-your-work
The easiest way to do this is to amend the last commit:
~~~console
`
	comment += fmt.Sprintf("$ git clone -b %q %s %s\n", pr.Head.Ref, pr.Head.Repo.SSHURL, "somewhere")
	comment += "$ cd somewhere\n"

	if pr.Commits > 1 {
		comment += fmt.Sprintf("$ git rebase -i HEAD~%d\n", pr.Commits)
		comment += "editor opens\nchange each 'pick' to 'edit'\nsave the file and quit\n"
	}

	comment += "$ git commit --amend -s --no-edit\n"
	if pr.Commits > 1 {
		comment += "$ git rebase --continue # and repeat the amend for each commit\n"
	}

	comment += "$ git push -f\n"
	comment += `~~~

This will update the existing PR, so **DO NOT** open a new one.
`

	return g.addUniqueComment(repo, strconv.Itoa(pr.Number), comment, "sign your commits", content)
}

func (g GitHub) removeComment(repo octokat.Repo, pr *octokat.PullRequest, commentType string, content *pullRequestContent) error {
	if c := content.FindComment(commentType, g.User); c != nil {
		return g.Client().RemoveComment(repo, c.Id)
	}

	return nil
}

func (g GitHub) addUniqueComment(repo octokat.Repo, prNum, comment, commentType string, content *pullRequestContent) error {
	// check if we already made the comment
	if content.AlreadyCommented(commentType, g.User) {
		return nil
	}

	// add the comment because we must not have already made it
	if _, err := g.Client().AddComment(repo, prNum, comment); err != nil {
		return err
	}

	log.Infof("Would have added comment about %q PR %s", commentType, prNum)
	return nil
}
