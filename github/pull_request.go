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

func (p *pullRequestContent) AlreadyCommented(commentType string) bool {
	for _, c := range p.comments {
		// if we already made the comment return nil
		if strings.ToLower(c.User.Login) == "gordontheturtle" && strings.Contains(c.Body, commentType) {
			log.Debugf("Already made comment about %q on PR %s", commentType, p.id)
			return true
		}
	}
	return false
}

func (p *pullRequestContent) FindComment(commentType string) *octokat.Comment {
	for _, c := range p.comments {
		if strings.ToLower(c.User.Login) == "gordontheturtle" && strings.Contains(c.Body, commentType) {
			return &c
		}
	}
	return nil
}

func (g *GitHub) getPullRequestContent(repo octokat.Repo, prId int) (*pullRequestContent, error) {
	var (
		files    []*octokat.PullRequestFile
		commits  []octokat.Commit
		comments []octokat.Comment
		err      error
	)
	n := strconv.Itoa(prId)

	if commits, err = g.Client().Commits(repo, n, &octokat.Options{}); err != nil {
		return nil, err
	}

	if files, err = g.Client().PullRequestFiles(repo, n, &octokat.Options{}); err != nil {
		return nil, err
	}

	if comments, err = g.Client().Comments(repo, n, &octokat.Options{}); err != nil {
		return nil, err
	}

	return &pullRequestContent{
		id:       prId,
		files:    files,
		commits:  commits,
		comments: comments,
	}, nil
}
