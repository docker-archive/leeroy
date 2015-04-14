package github

import (
	"testing"

	"github.com/crosbymichael/octokat"
)

type prExpected struct {
	expected bool
	pr       *octokat.PullRequest
}

func TestIsMergeable(t *testing.T) {
	f := false
	y := true
	prs := []prExpected{
		prExpected{
			true,
			&octokat.PullRequest{
				Mergeable: nil,
			},
		},
		prExpected{
			true,
			&octokat.PullRequest{
				Mergeable: &y,
			},
		},
		prExpected{
			false,
			&octokat.PullRequest{
				Mergeable: &f,
			},
		},
	}

	for _, pe := range prs {
		if pe.expected != isMergeable(pe.pr) {
			t.Fatalf("expected %v, was %v, for: %#v\n", pe.expected, isMergeable(pe.pr), pe)
		}
	}
}
