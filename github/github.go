package github

import (
	"net/http"
	"os"

	"github.com/crosbymichael/octokat"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
)

// GitHub holds the client information for connecting to the GitHub API
type GitHub struct {
	AuthToken string
	User      string
}

// Client initializes the authorization with the GitHub API
func (g GitHub) Client() *octokat.Client {
	var cache httpcache.Cache
	if cachePath := os.Getenv("GITHUB_CACHE_PATH"); cachePath != "" {
		cache = diskcache.New(cachePath)
	} else {
		cache = httpcache.NewMemoryCache()
	}
	tr := httpcache.NewTransport(cache)

	c := &http.Client{Transport: tr}

	gh := octokat.NewClient()
	gh = gh.WithToken(g.AuthToken)
	gh = gh.WithHTTPClient(c)
	return gh
}

func nameWithOwner(repo *octokat.Repository) octokat.Repo {
	return octokat.Repo{
		Name:     repo.Name,
		UserName: repo.Owner.Login,
	}
}

func bot(user octokat.User) bool {
	return user.Login == "GordonTheTurtle"
}
