package github

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

func (g GitHub) addDCOUnsignedComment(repo octokat.Repo, pr *PullRequest, content *PullRequestContent) error {
	comment := `Please sign your commits following these rules:
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

Ammending updates the existing PR. You **DO NOT** need to open a new one.
`

	return g.addUniqueComment(repo, strconv.Itoa(pr.Number), comment, "sign your commits", content)
}

func (g GitHub) addNeedMoreInfoComment(repo octokat.Repo, issueNum int, content *PullRequestContent) error {
	comment := `
If you are reporting a new issue, make sure that we do not have any duplicates already open. You can ensure this by searching the issue list for this repository. If there is a duplicate, please close your issue and add a comment to the existing issue instead.

If you suspect your issue is a bug, please edit your issue description to include the BUG REPORT INFORMATION shown below. If you fail to provide this information within 7 days, we cannot debug your issue and will close it. We will, however, reopen it if you later provide the information.

For more information about reporting issues, see [CONTRIBUTING.md](https://github.com/docker/docker/blob/master/CONTRIBUTING.md#reporting-other-issues).

**You _don't_ have to include this information if this is a _feature request_**

(This is an automated, informational response)

-----------------------------
BUG REPORT INFORMATION
-----------------------------
Use the commands below to provide key information from your environment:

` + "`docker version`" + `:
` + "`docker info`" + `:

Provide additional environment details (AWS, VirtualBox, physical, etc.):



List the steps to reproduce the issue:
1.
2.
3.


Describe the results you received:


Describe the results you expected:


Provide additional info you think is important:


----------END REPORT ---------



#ENEEDMOREINFO
`

	return g.addUniqueComment(repo, strconv.Itoa(issueNum), comment, "#ENEEDMOREINFO", content)
}

func (g GitHub) removeComment(repo octokat.Repo, commentType string, content *PullRequestContent) error {
	if c := content.FindComment(commentType, g.User); c != nil {
		return g.Client().RemoveComment(repo, c.Id)
	}

	return nil
}

func (g GitHub) addUniqueComment(repo octokat.Repo, prNum, comment, commentType string, content *PullRequestContent) error {
	// check if we already made the comment
	if content.AlreadyCommented(commentType, g.User) {
		return nil
	}

	// add the comment because we must not have already made it
	if _, err := g.Client().AddComment(repo, prNum, comment); err != nil {
		return err
	}

	logrus.Infof("Added comment about %q PR/issue %s", commentType, prNum)
	return nil
}
