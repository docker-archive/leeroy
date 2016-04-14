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
		{
			true,
			&octokat.PullRequest{
				Mergeable: nil,
			},
		},
		{
			true,
			&octokat.PullRequest{
				Mergeable: &y,
			},
		},
		{
			false,
			&octokat.PullRequest{
				Mergeable: &f,
			},
		},
	}

	for _, pe := range prs {
		p := &PullRequest{PullRequest: pe.pr}
		mergeable := isMergeable(p, "")
		if pe.expected != mergeable {
			t.Fatalf("expected %v, was %v, for: %#v\n", pe.expected, mergeable, pe)
		}
	}
}
