package github

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/crosbymichael/octokat"
)

var (
	dcoRegex = regexp.MustCompile("(?m)(Docker-DCO-1.1-)?Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>( \\(github: ([a-zA-Z0-9][a-zA-Z0-9-]+)\\))?")
)

type pullRequestContent struct {
	files   []*octokat.PullRequestFile
	commits []octokat.Commit
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

func (g *GitHub) getPullRequestContent(repo octokat.Repo, prId int) (*pullRequestContent, error) {
	var (
		errs    []error
		wg      sync.WaitGroup
		commits []octokat.Commit
		files   []*octokat.PullRequestFile
	)

	n := strconv.Itoa(prId)

	wg.Add(2)
	go func() {
		defer wg.Done()

		var err error
		commits, err = g.Client().Commits(repo, n, nil)
		if err != nil {
			errs = append(errs, err)
		}
	}()

	go func() {
		defer wg.Done()

		var err error
		files, err = g.Client().PullRequestFiles(repo, n, nil)
		if err != nil {
			errs = append(errs, err)
		}
	}()

	wg.Wait()

	if len(errs) > 0 {
		b := bytes.NewBufferString("")
		for _, e := range errs {
			b.WriteString(e.Error())
		}

		return nil, errors.New(b.String())
	}

	return &pullRequestContent{files, commits}, nil
}
