#!/bin/bash
set -e

: ${LEEROY_URI:=https://leeroy.dockerproject.org}
URI=https://api.github.com
API_VERSION=v3

if [[ -z "$GITHUB_TOKEN" ]]; then
	echo "Set the GITHUB_TOKEN env variable."
	return 1
fi

if [[ -z "$LEEROY_AUTH" ]]; then
	echo "Set the LEEROY_AUTH env variable."
	return 1
fi

API_HEADER="Accept: application/vnd.github.${API_VERSION}+json"
AUTH_HEADER="Authorization: token ${GITHUB_TOKEN}"

DEFAULT_PER_PAGE=10

rebuild_pulls(){
	local repo=$1
	local page=$2

	if [[ -z "$repo" ]]; then
		echo "Pass a repo as the first arguement, ex. docker/docker"
		return 1
	fi

	if [[ -z "$page" ]]; then
		page=1
	fi

	# send the request
	local response=$(curl -i -sSL -H "${AUTH_HEADER}" -H "${API_HEADER}" ${URI}/repos/${repo}/pulls?per_page=${DEFAULT_PER_PAGE}&page=1)

	# seperate the headers and body into 2 variables
	head=true
	while read -r line; do 
		if $head; then 
			if [[ $line = $'\r' ]]; then
				head=false
			else
				header="$header"$'\n'"$line"
			fi
		else
			body="$body"$'\n'"$line"
		fi
	done < <(echo "$response")

	statuses=$(echo $body | jq --raw-output '.[] | {id: .number, uri: .statuses_url} | tostring')

	for s in $statuses; do
		id=$(echo $s | jq --raw-output '.id')
		uri=$(echo $s | jq --raw-output '.uri')

		# send the request
		response=$(curl -sSL -H "${AUTH_HEADER}" -H "${API_HEADER}" ${uri})

		# find the statuses that are not success
		contexts=$(echo $response | jq --raw-output '.[] | select(.state != "success") | .context')
		for context in $contexts; do
			echo "Rebuilding pull request ${id} for context $context"
			data='{"repo":"'${repo}'","context":"'${context}'"}'
			echo "curl -ssL -X POST -d "${data}" ${LEEROY_URI}/build/cron"
		done
	done

	if [[ "$page" == "1" ]]; then
		# get the last page from the headers
		header=${header%%" rel=\"last\""*}
		header=${header#*"rel=\"next\""}
		header=${header%%">;"*}
		last_page=${header#*"&page="}
		echo last page is $last_page
		if [[ ! -z "$last_page" ]] || [[ "$last_page" != "$page" ]]; then
			for page in  $(seq $((page + 1)) 1 ${last_page}); do
				rebuild_pulls $repo $page
			done
		fi
	fi
}

rebuild_pulls docker/docker
