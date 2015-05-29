package github

import (
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

var (
	dcoRegex = regexp.MustCompile("(?m)(Docker-DCO-1.1-)?Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>( \\(github: ([a-zA-Z0-9][a-zA-Z0-9-]+)\\))?")
)

type pullRequestContent struct {
	id       int
	files    []*octokat.PullRequestFile
	commits  []octokat.Commit
	comments []octokat.Comment
}

func (p *pullRequestContent) IsDocsOnly() bool {
	if len(p.files) == 0 {
		return false
	}

	for _, f := range p.files {
		if !strings.HasSuffix(f.FileName, ".md") || !strings.HasPrefix(f.FileName, "docs") {
			return false
		}
	}

	return true
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
			log.Debugf("Already made comment about %q on PR %s", commentType, p.id)
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
		if strings.HasSuffix(f.FileName, "_freebsd.go") {
			freebsd = true
		} else if strings.HasSuffix(f.FileName, "_linux.go") {
			linux = true
		}
	}

	return freebsd && !linux
}

func (p *pullRequestContent) OnlyWindows() bool {
	var windows bool
	var linux bool

	for _, f := range p.files {
		if strings.HasSuffix(f.FileName, "_windows.go") {
			windows = true
		} else if strings.HasSuffix(f.FileName, "_linux.go") {
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

	if isPR {
		if commits, err = g.Client().Commits(repo, n, &octokat.Options{}); err != nil {
			return nil, err
		}

		if files, err = g.Client().PullRequestFiles(repo, n, &octokat.Options{}); err != nil {
			return nil, err
		}
	}

	if comments, err = g.Client().Comments(repo, n, &octokat.Options{}); err != nil {
		return nil, err
	}

	return &pullRequestContent{
		id:       id,
		files:    files,
		commits:  commits,
		comments: comments,
	}, nil
}
