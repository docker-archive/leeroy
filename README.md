## Leeroy: Jenkins integration with GitHub pull requests

[![Circle CI](https://circleci.com/gh/jfrazelle/leeroy.svg?style=svg)](https://circleci.com/gh/jfrazelle/leeroy)

Leeroy is a Go application which integrates Jenkins with 
GitHub pull requests.  
Leeroy uses [GitHub hooks](http://developer.github.com/v3/repos/hooks/) 
to listen for pull request notifications and starts jobs on your Jenkins 
server.  Using the Jenkins [notification plugin][jnp], Leeroy updates the 
pull request using GitHub's 
[status API](http://developer.github.com/v3/repos/statuses/)
with pending, success, failure, or error statuses.

### Configuration

Leeroy needs to be configured to point to your GitHub repositories,
to your Jenkins server and its jobs.  You will need to add a GitHub 
webook pointing towards your leeroy instance at the endpoint 
`/notifications/github`. You will also need to configure your
Jenkins jobs to pull the right repositories and commits.

#### Leeroy Configuration

Below is a sample leeroy config file:

```
{
    "jenkins": {
        "username": "leeroy",
        "token": "YOUR_JENKINS_API_TOKEN",
        "base_url": "https://jenkins.dockerproject.com"
    },
    
    // Whether a Jenkins job is created for each commit in a pull request,
    // or only one for the last one.
    // What commits to build in a pull request. There are three options:
    // "all": build all commits in the pull request.
    // "last": build only the last commit in the pull request.
    // "new": build only commits that don't already have a commit status set.
    "build_commits": "last", // (default)
    
    "github_token": "YOUR_GITHUB_TOKEN",
    
    // A list of dicts containing configuration for each GitHub repository &
    // Jenkins job pair you want to join together.
    "builds": [
        {
            "github_repo": "docker/docker",
            "jenkins_job_name": "Docker-PRs",
            "context": "janky" // context to send to github for status (if you
            wanna stack em)
        }
    ],

    // Basic Auth for endoints
    "user": "USER",
    "pass": "PASS"
}
```

#### Jenkins Configuration

1. Install the Jenkins [git plugin][jgp] and [notification plugin][jnp].

2. Create a Jenkins job.  Under "Job Notifications", set a Notification
Endpoint with protocol HTTP and the URL pointing to `/notification/jenkins`
on your Leeroy server.  If your Leeroy server is `leeroy.example.com`, set
this to `http://leeroy.example.com/notification/jenkins`.

3. Check the "This build is parameterized" checkbox, and add 4 string
parameters: `GIT_BASE_REPO`, `GIT_HEAD_REPO`, `GIT_SHA1`, and `GITHUB_URL`.
Default values like `username/repo` for `GIT_BASE_REPO` and `GIT_HEAD_REPO`,
and `master` for `GIT_SHA1` are a good idea, but not required.

4. Under "Source Code Management", select Git.  Set the "Repository URL" to
`git@github.com:$GIT_HEAD_REPO.git`.  Set "Branch Specifier" to `$GIT_SHA1`.

5. Configure the rest of the job however you would otherwise.

[jgp]: https://wiki.jenkins-ci.org/display/JENKINS/Git+Plugin
[jnp]: https://wiki.jenkins-ci.org/display/JENKINS/Notification+Plugin


### Usage

```console
$ leeroy -h
Usage of leeroy:
  -cert="": path to ssl certificate
  -config="/etc/leeroy/config.json": path to config file
  -d=false: run in debug mode
  -key="": path to ssl key
  -port="80": port to use
  -v=false: print version and exit (shorthand)
  -version=false: print version and exit
```


