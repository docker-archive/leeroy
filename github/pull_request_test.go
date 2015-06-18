package github

import (
	"testing"

	"github.com/crosbymichael/octokat"
)

func TestCommitsAreSigned(t *testing.T) {
	cases := map[string]bool{
		"":                                                                                                   false,
		"Signed-off-by:":                                                                                     false,
		"Signed-off-by: David Calavera <david.calavera@gmail.com>":                                           true,
		"\n\nSigned-off-by: David Calavera <david.calavera@gmail.com>\n\n":                                   true,
		"\n\nDocker-DCO-1.1-Signed-off-by: David Calavera <david.calavera@gmail.com> (github: calavera)\n\n": true,
	}

	for message, valid := range cases {
		commits := []octokat.Commit{
			octokat.Commit{
				Commit: &octokat.CommitCommit{
					Message: message,
				},
			},
		}

		pr := &pullRequestContent{commits: commits}
		s := pr.CommitsSigned()

		if s != valid {
			t.Fatalf("expected %v, was %v, for: %s\n", valid, s, message)
		}
	}
}

func TestFilesAreDocs(t *testing.T) {
	cases := []struct {
		files []string
		valid bool
	}{
		{[]string{""}, false},
		{[]string{"file.md"}, true},
		{[]string{"docs/file.txt"}, true},
		{[]string{"docs/file.md"}, true},
		{[]string{"docs/hub/file.md"}, true},
		{[]string{"man/file.txt"}, true},
		{[]string{"docs/hub/file.md", "man/file.txt"}, true},
		{[]string{"experimental/file.txt"}, true},
		{[]string{"daemon/daemon.go", "experimental/file.txt"}, false},
	}

	for _, c := range cases {
		var files []*octokat.PullRequestFile
		for _, f := range c.files {
			files = append(files, &octokat.PullRequestFile{
				FileName: f,
			})
		}

		pr := &pullRequestContent{files: files}
		s := pr.IsDocsOnly()

		if s != c.valid {
			t.Fatalf("expected %v, was %v, for: %s\n", c.valid, s, c.files)
		}
	}
}

func TestAlreadyCommented(t *testing.T) {
	cases := []struct {
		login string
		body  string
		exp   bool
	}{
		{
			"calavera",
			"sign your commits",
			false,
		},
		{
			"calavera",
			":+1:",
			false,
		},
		{
			"gordontheturtle",
			"rebase your commits",
			false,
		},
		{
			"gordontheturtle",
			"sign your commits",
			true,
		},
	}

	for _, c := range cases {
		comments := []octokat.Comment{
			octokat.Comment{
				User: octokat.User{
					Login: c.login,
				},
				Body: c.body,
			},
		}

		pr := &pullRequestContent{comments: comments}
		if done := pr.AlreadyCommented("sign your commits", "gordontheturtle"); done != c.exp {
			t.Fatalf("Expected commented %v, but was %v for user %s and body %s\n", c.exp, done, c.login, c.body)
		}

		comment := pr.FindComment("sign your commits", "gordontheturtle")
		found := comment != nil

		if found != c.exp {
			t.Fatalf("Expected found %v, but was %v for user %s and body %s\n", c.exp, found, c.login, c.body)
		}
	}
}
