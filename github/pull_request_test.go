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
			{
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

func TestDistribution(t *testing.T) {
	cases := []struct {
		files []string
		valid bool
	}{
		{[]string{""}, false},
		{[]string{"file.md"}, false},
		{[]string{"docs/file.txt"}, false},
		{[]string{"graph/file.go"}, true},
		{[]string{"registry/file.go"}, true},
		{[]string{"image/file.go"}, true},
		{[]string{"trust/file.go"}, true},
		{[]string{"builder/file.go"}, true},
		{[]string{"something/with/builder/file.go"}, false},
	}

	for _, c := range cases {
		var files []*octokat.PullRequestFile
		for _, f := range c.files {
			files = append(files, &octokat.PullRequestFile{
				FileName: f,
			})
		}

		pr := &pullRequestContent{files: files}
		s := pr.Distribution()

		if s != c.valid {
			t.Fatalf("expected %v, was %v, for: %s\n", c.valid, s, c.files)
		}
	}
}

func TestHasVendoringChanges(t *testing.T) {
	cases := []struct {
		files []string
		valid bool
	}{
		{[]string{""}, false},
		{[]string{"file.md"}, false},
		{[]string{"docs/file.txt"}, false},
		{[]string{"vendor/anything.really"}, true},
		{[]string{"hack/vendor.sh"}, true},
		{[]string{"hack/.vendor-helper.sh"}, true},
		{[]string{"hack/.vendor-helper.sh", "daemon/daemon.go"}, true},
	}

	for _, c := range cases {
		var files []*octokat.PullRequestFile
		for _, f := range c.files {
			files = append(files, &octokat.PullRequestFile{
				FileName: f,
			})
		}

		pr := &pullRequestContent{files: files}
		s := pr.HasVendoringChanges()

		if s != c.valid {
			t.Fatalf("expected %v, was %v, for: %s\n", c.valid, s, c.files)
		}
	}
}

func TestHasDocsChanges(t *testing.T) {
	cases := []struct {
		files []string
		valid bool
	}{
		{[]string{""}, false},
		{[]string{"file.md"}, false},
		{[]string{"docs/file.txt"}, true},
		{[]string{"docs/file.md"}, true},
		{[]string{"docs/hub/file.md"}, true},
		{[]string{"man/file.txt"}, false},
		{[]string{"docs/hub/file.md", "man/file.txt"}, true},
		{[]string{"experimental/file.txt"}, false},
		{[]string{"daemon/daemon.go", "experimental/file.txt"}, false},
		{[]string{"daemon/README.txt", "experimental/file.txt"}, false},
	}

	for _, c := range cases {
		var files []*octokat.PullRequestFile
		for _, f := range c.files {
			files = append(files, &octokat.PullRequestFile{
				FileName: f,
			})
		}

		pr := &pullRequestContent{files: files}
		s := pr.HasDocsChanges()

		if s != c.valid {
			t.Fatalf("expected %v, was %v, for: %s\n", c.valid, s, c.files)
		}
	}
}

func TestIsNonCodeOnly(t *testing.T) {
	cases := []struct {
		files []string
		valid bool
	}{
		{[]string{""}, false},
		{[]string{"file.md"}, false},
		{[]string{"man/file.md"}, true},
		{[]string{"man/file.txt"}, true},
		{[]string{"docs/file.md"}, true},
		{[]string{"man/file.sh"}, false},
		{[]string{"docs/file.go"}, true},
		{[]string{"docs/hub/file.md"}, true},
		{[]string{"docs/hub/file.md", "man/file.txt"}, true},
		{[]string{"experimental/file.md"}, true},
		{[]string{"experimental/file.qo"}, true},
		{[]string{"contrib/completion/zsh/_docker"}, true},
		{[]string{"contrib/desktop-integration/README.md"}, true},
		{[]string{"contrib/mkimage-alpine.sh"}, true},
		{[]string{"docs/file.txt", "daemon/daemon.go"}, false},
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
		s := pr.IsNonCodeOnly()

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
			{
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
