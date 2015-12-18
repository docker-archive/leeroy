package github

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

var (
	dcoRegex = regexp.MustCompile("(?m)(Docker-DCO-1.1-)?Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>( \\(github: ([a-zA-Z0-9][a-zA-Z0-9-]+)\\))?")
)

type PullRequest struct {
	Hook    *octokat.PullRequestHook
	Repo    octokat.Repo
	Content *pullRequestContent
	*octokat.PullRequest
}

func (g GitHub) LoadPullRequest(hook *octokat.PullRequestHook) (*PullRequest, error) {
	pr := hook.PullRequest
	repo := nameWithOwner(hook.Repo)

	content, err := g.getContent(repo, hook.Number, true)
	if err != nil {
		return nil, err
	}

	return &PullRequest{
		Hook:        hook,
		Repo:        repo,
		Content:     content,
		PullRequest: pr,
	}, nil
}

func (pr PullRequest) ReleaseBase() bool {
	return pr.Base.Ref == "release"
}

type pullRequestContent struct {
	id       int
	files    []*octokat.PullRequestFile
	commits  []octokat.Commit
	comments []octokat.Comment
}

func (p *pullRequestContent) HasDocsChanges() bool {
	if len(p.files) == 0 {
		return false
	}

	// Did any files in the docs dir change?
	for _, f := range p.files {
		if strings.HasPrefix(f.FileName, "docs") {
			return true
		}
	}

	return false
}

func (p *pullRequestContent) IsNonCodeOnly() bool {
	if len(p.files) == 0 {
		return false
	}

	// if there are any changed files not in docs/man/experimental dirs
	for _, f := range p.files {
		if !strings.HasSuffix(f.FileName, ".md") &&
			!strings.HasSuffix(f.FileName, ".txt") &&
			!strings.HasPrefix(f.FileName, "docs") &&
			!strings.HasPrefix(f.FileName, "man") &&
			!strings.HasPrefix(f.FileName, "experimental") {
			return false
		}
	}

	return true
}

func (p *pullRequestContent) Distribution() bool {
	if len(p.files) == 0 {
		return false
	}

	for _, f := range p.files {
		if anyPackage(f.FileName, "registry", "graph", "image", "trust", "builder") {
			return true
		}
	}

	return false
}

func (p *pullRequestContent) CommitsSigned() bool {
	if len(p.commits) == 0 {
		return true
	}

	for _, c := range p.commits {
		if !dcoRegex.MatchString(c.Commit.Message) {
			return false
		}
	}

	return true
}

func (p *pullRequestContent) AlreadyCommented(commentType, user string) bool {
	for _, c := range p.comments {
		// if we already made the comment return nil
		if strings.ToLower(c.User.Login) == user && strings.Contains(c.Body, commentType) {
			logrus.Debugf("Already made comment about %q on PR %s", commentType, p.id)
			return true
		}
	}
	return false
}

func (p *pullRequestContent) FindComment(commentType, user string) *octokat.Comment {
	for _, c := range p.comments {
		if strings.ToLower(c.User.Login) == user && strings.Contains(c.Body, commentType) {
			return &c
		}
	}
	return nil
}

func (p *pullRequestContent) OnlyFreebsd() bool {
	var freebsd bool
	var linux bool

	for _, f := range p.files {
		if strings.HasPrefix(f.FileName, "_freebsd.go") {
			freebsd = true
		} else if strings.HasPrefix(f.FileName, "_linux.go") {
			linux = true
		}
	}

	return freebsd && !linux
}

func (p *pullRequestContent) OnlyWindows() bool {
	var windows bool
	var linux bool

	for _, f := range p.files {
		if strings.HasPrefix(f.FileName, "_windows.go") {
			windows = true
		} else if strings.HasPrefix(f.FileName, "_linux.go") {
			linux = true
		}
	}

	return windows && !linux
}

func (g *GitHub) getContent(repo octokat.Repo, id int, isPR bool) (*pullRequestContent, error) {
	var (
		files    []*octokat.PullRequestFile
		commits  []octokat.Commit
		comments []octokat.Comment
		err      error
	)
	n := strconv.Itoa(id)

	options := &octokat.Options{
		QueryParams: map[string]string{"per_page": "100"},
	}

	if isPR {
		if commits, err = g.Client().Commits(repo, n, options); err != nil {
			return nil, err
		}

		if files, err = g.Client().PullRequestFiles(repo, n, options); err != nil {
			return nil, err
		}
	}

	if comments, err = g.Client().Comments(repo, n, options); err != nil {
		return nil, err
	}

	return &pullRequestContent{
		id:       id,
		files:    files,
		commits:  commits,
		comments: comments,
	}, nil
}

func anyPackage(fileName string, packages ...string) bool {
	for _, p := range packages {
		if strings.HasPrefix(fileName, p) {
			return true
		}
	}
	return false
}
