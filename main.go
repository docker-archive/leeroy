package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/leeroy/jenkins"
)

const (
	// VERSION is the version
	VERSION = "v0.1.0"
	// DEFAULTCONTEXT is the default github context for a build
	DEFAULTCONTEXT = "janky"
)

var (
	certFile   string
	keyFile    string
	port       string
	configFile string
	debug      bool
	version    bool

	config Config = Config{CheckDCO: true}
)

// Config describes the leeroy config file
type Config struct {
	Jenkins      jenkins.Client `json:"jenkins"`
	BuildCommits string         `json:"build_commits"`
	GHToken      string         `json:"github_token"`
	GHUser       string         `json:"github_user"`
	Builds       []Build        `json:"builds"`
	User         string         `json:"user"`
	Pass         string         `json:"pass"`
	CheckDCO     bool           `json:"check_dco"`
}

// Build describes the paramaters for a build
type Build struct {
	Repo         string `json:"github_repo"`
	Job          string `json:"jenkins_job_name"`
	Context      string `json:"context"`
	Custom       bool   `json:"custom"`
	HandleIssues bool   `json:"handle_issues"`
	IsPipeline   bool   `json:"is_pipeline"`
}

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")
	flag.StringVar(&certFile, "cert", "", "path to ssl certificate")
	flag.StringVar(&keyFile, "key", "", "path to ssl key")
	flag.StringVar(&port, "port", "80", "port to use")
	flag.StringVar(&configFile, "config", "/etc/leeroy/config.json", "path to config file")
	flag.Parse()
}

func main() {
	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if version {
		fmt.Println(VERSION)
		return
	}

	// read the config file
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		logrus.Errorf("config file does not exist: %s", configFile)
		return
	}
	c, err := ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Errorf("could not read config file: %v", err)
		return
	}
	if err := json.Unmarshal(c, &config); err != nil {
		logrus.Errorf("error parsing config file as json: %v", err)
		return
	}

	// create mux server
	mux := http.NewServeMux()

	// ping endpoint
	mux.HandleFunc("/ping", pingHandler)

	// jenkins notification endpoint
	mux.HandleFunc("/notification/jenkins", jenkinsHandler)

	// github webhooks endpoint
	mux.HandleFunc("/notification/github", githubHandler)

	// retry build endpoint
	mux.HandleFunc("/build/retry", customBuildHandler)

	// custom build endpoint
	mux.HandleFunc("/build/custom", customBuildHandler)

	// cron endpoint to reschedule bulk jobs
	mux.HandleFunc("/build/cron", cronBuildHandler)

	// set up the server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	logrus.Printf("Starting server on port %q", port)
	if certFile != "" && keyFile != "" {
		logrus.Fatal(server.ListenAndServeTLS(certFile, keyFile))
	} else {
		logrus.Fatal(server.ListenAndServe())
	}
}
