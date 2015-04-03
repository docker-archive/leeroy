package github

import "github.com/crosbymichael/octokat"

type GitHub struct {
	AuthToken string
}

func (g GitHub) Client() *octokat.Client {
	gh := octokat.NewClient()
	gh = gh.WithToken(g.AuthToken)
	return gh
}
