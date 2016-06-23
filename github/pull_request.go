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

// PullRequest describes a github pull request
type PullRequest struct {
	Hook    *octokat.PullRequestHook
	Repo    octokat.Repo
	Content *PullRequestContent
	*octokat.PullRequest
}

// LoadPullRequest takes an incoming PullRequestHook and converts it to the PullRequest type
func (g GitHub) LoadPullRequest(hook *octokat.PullRequestHook) (*PullRequest, error) {
	pr := hook.PullRequest
	repo := nameWithOwner(hook.Repo)

	content, err := g.GetContent(repo, hook.Number, true)
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

// ReleaseBase checks if the pull request is being merged into the release branch
func (pr PullRequest) ReleaseBase() bool {
	return pr.Base.Ref == "release"
}

// Execdriver checks if the changes are to execdriver's directories.
func (pr *PullRequest) Execdriver() bool {
	if len(pr.Content.files) == 0 || strings.Contains(strings.ToLower(pr.Title), "containerd") {
		return false
	}

	for _, f := range pr.Content.files {
		if anyPackage(f.FileName, "daemon/execdriver") {
			return true
		}
	}

	return false
}

// PullRequestContent contains the files, commits, and comments for a given
// pull request
type PullRequestContent struct {
	id       int
	files    []*octokat.PullRequestFile
	commits  []octokat.Commit
	comments []octokat.Comment
}

// HasVendoringChanges checks for vendoring changes.
func (p *PullRequestContent) HasVendoringChanges() bool {
	if len(p.files) == 0 {
		return false
	}

	// Did any files in the vendor dir change?
	for _, f := range p.files {
		if isVendor(f.FileName) {
			return true
		}
	}

	return false
}

// HasDocsChanges checks for docs changes.
func (p *PullRequestContent) HasDocsChanges() bool {
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

// IsNonCodeOnly chacks if only non code files are modified.
func (p *PullRequestContent) IsNonCodeOnly() bool {
	if len(p.files) == 0 {
		return false
	}

	// if there are any changed files not in docs/man/experimental dirs
	for _, f := range p.files {
		if !isMan(f.FileName) &&
			!isDocumentation(f.FileName) &&
			!isExperimental(f.FileName) &&
			!isContrib(f.FileName) {
			return false
		}
	}

	return true
}

// Distribution checks if the changes are to distribution's directories.
func (p *PullRequestContent) Distribution() bool {
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

// CommitsSigned checks if the commits are signed.
func (p *PullRequestContent) CommitsSigned() bool {
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

// AlreadyCommented checks if the user has already commented.
func (p *PullRequestContent) AlreadyCommented(commentType, user string) bool {
	for _, c := range p.comments {
		// if we already made the comment return nil
		if strings.ToLower(c.User.Login) == user && strings.Contains(c.Body, commentType) {
			logrus.Debugf("Already made comment about %q on PR %s", commentType, p.id)
			return true
		}
	}
	return false
}

// FindComment finds a specific comment.
func (p *PullRequestContent) FindComment(commentType, user string) *octokat.Comment {
	for _, c := range p.comments {
		if strings.ToLower(c.User.Login) == user && strings.Contains(c.Body, commentType) {
			return &c
		}
	}
	return nil
}

// OnlyFreebsd checks if changes are only to freebsd specific files.
func (p *PullRequestContent) OnlyFreebsd() bool {
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

// OnlyWindows checks if changes are only to windows specific files.
func (p *PullRequestContent) OnlyWindows() bool {
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

// Protobuf checks if there are changes to protocol buffers definitions or
// code generated from them.
func (p *PullRequestContent) Protobuf() bool {
	for _, f := range p.files {
		if strings.HasSuffix(f.FileName, ".proto") || strings.HasSuffix(f.FileName, ".pb.go") {
			return true
		}
	}

	return false
}

// GetContent returns the content of the issue/pull request number passed.
func (g *GitHub) GetContent(repo octokat.Repo, id int, isPR bool) (*PullRequestContent, error) {
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

	return &PullRequestContent{
		id:       id,
		files:    files,
		commits:  commits,
		comments: comments,
	}, nil
}

func anyPackage(fileName string, packages ...string) bool {
	return hasAny(strings.HasPrefix, fileName, packages...)
}

func isMan(filename string) bool {
	return hasAny(strings.HasPrefix, filename, "man") &&
		hasAny(strings.HasSuffix, filename, ".md", ".txt")
}

func isDocumentation(filename string) bool {
	return hasAny(strings.HasPrefix, filename, "docs")
}

func isExperimental(filename string) bool {
	return hasAny(strings.HasPrefix, filename, "experimental")
}

func isContrib(filename string) bool {
	contribs := []string{
		"contrib/completion",
		"contrib/desktop-integration",
		"contrib/mkimage",
	}
	return hasAny(strings.HasPrefix, filename, contribs...)
}

func isVendor(filename string) bool {
	return hasAny(strings.HasPrefix, filename, "vendor", "hack/vendor.sh", "hack/.vendor-helper.sh")
}

func hasAny(fn func(string, string) bool, s string, cases ...string) bool {
	for _, c := range cases {
		if fn(s, c) {
			return true
		}
	}
	return false
}
