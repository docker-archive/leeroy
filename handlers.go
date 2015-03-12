package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
	"github.com/jfrazelle/leeroy/jenkins"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
	return
}

func jenkinsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Errorf("%q is not a valid method", r.Method)
		w.WriteHeader(405)
		return
	}

	// decode the body
	decoder := json.NewDecoder(r.Body)
	var j jenkins.JenkinsResponse
	if err := decoder.Decode(&j); err != nil {
		log.Errorf("decoding the jenkins request as json failed: %v", err)
		return
	}

	log.Infof("Received Jenkins notification for %s %d (%s): %s", j.Name, j.Build.Number, j.Build.Url, j.Build.Phase)

	// if the phase is not started or completed
	// we don't care
	if j.Build.Phase != "STARTED" && j.Build.Phase != "COMPLETED" {
		return
	}

	// get the status for github
	// and create a status description
	desc := fmt.Sprintf("Jenkins build %s %d", j.Name, j.Build.Number)
	var state string
	if j.Build.Phase == "STARTED" {
		state = "pending"
		desc += " is running"
	} else {

		switch j.Build.Status {
		case "SUCCESS":
			state = "success"
			desc += " has succeeded"
		case "FAILURE":
			state = "failure"
			desc += " has failed"
		case "UNSTABLE":
			state = "failure"
			desc += " was unstable"
		case "ABORTED":
			state = "error"
			desc += " has encountered an error"
		default:
			log.Errorf("Did not understand %q build status. Aborting.", j.Build.Status)
			return
		}
	}
	// get the build
	build, err := config.getBuildByJob(j.Name)
	if err != nil {
		log.Error(err)
		return
	}

	// update the github status
	if err := config.updateGithubStatus(j.Build.Parameters.GitBaseRepo, build.Context, j.Build.Parameters.GitSha, state, desc, j.Build.Url+"console"); err != nil {
		log.Error(err)
	}

	return
}

func githubHandler(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-GitHub-Event")

	switch event {
	case "":
		log.Error("Got GitHub notification without a type")
		return
	case "ping":
		w.WriteHeader(200)
		return
	case "pull_request":
		log.Debugf("Got a pull request hook")
	default:
		fmt.Errorf("Got unknown GitHub notification event type: %s", event)
		return
	}

	// parse the pull request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Error reading github handler body: %v", err)
		w.WriteHeader(500)
		return
	}
	prHook, err := octokat.ParsePullRequestHook(body)
	if err != nil {
		log.Errorf("Error parsing hook: %v", err)
		w.WriteHeader(500)
		return
	}

	pr := prHook.PullRequest
	baseRepo := fmt.Sprintf("%s/%s", pr.Base.Repo.Owner.Login, pr.Base.Repo.Name)

	log.Infof("Received GitHub pull request notification for %s %d (%s): %s", baseRepo, pr.Number, pr.URL, prHook.Action)

	// ignore everything we don't care about
	if prHook.Action != "opened" && prHook.Action != "reopened" && prHook.Action != "synchronize" {
		log.Debugf("Ignoring PR hook action %q", prHook.Action)
		return
	}

	// get the builds
	builds, err := config.getBuilds(baseRepo, false)
	if err != nil {
		log.Error(err)
		w.WriteHeader(500)
		return
	}

	// schedule the jenkins builds
	for _, build := range builds {
		if err := config.scheduleJenkinsBuild(baseRepo, pr.Number, build); err != nil {
			log.Error(err)
			w.WriteHeader(500)
		}
	}

	return
}

type requestBuild struct {
	Number  int    `json:"number"`
	Repo    string `json:"repo"`
	Context string `json:"context"`
}

func customBuildHandler(w http.ResponseWriter, r *http.Request) {
	// setup auth
	user, pass, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(401)
		return
	}
	if user != config.User && pass != config.Pass {
		w.WriteHeader(401)
		return
	}

	if r.Method != "POST" {
		fmt.Errorf("%q is not a valid method", r.Method)
		w.WriteHeader(405)
		return
	}

	// decode the body
	decoder := json.NewDecoder(r.Body)
	var b requestBuild
	if err := decoder.Decode(&b); err != nil {
		log.Errorf("decoding the retry request as json failed: %v", err)
		w.WriteHeader(500)
		return
	}

	// get the build
	build, err := config.getBuildByContext(b.Context)
	if err != nil {
		log.Error(err)
		w.WriteHeader(500)
		return
	}

	// schedule the jenkins build
	if err := config.scheduleJenkinsBuild(b.Repo, b.Number, build); err != nil {
		w.WriteHeader(500)
		log.Error(err)
		return
	}

	w.WriteHeader(204)
	return
}
