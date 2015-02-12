# Leeroy


### Sample Config File

`/etc/leeroy/config.json`

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
    "build_commits": "new", // (default)
    
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
    ]
}
```
