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
LAST_PAGE=1

# get the last page from the headers
get_last_page(){
	header=${1%%" rel=\"last\""*}
	header=${header#*"rel=\"next\""}
	header=${header%%">;"*}
	LAST_PAGE=$(echo ${header#*"&page="} | bc 2>/dev/null)
}

rebuild_pulls(){
	local repo=$1
	local page=$2

	# send the request
	local response=$(curl -i -sSL -H "${AUTH_HEADER}" -H "${API_HEADER}" "${URI}/repos/${repo}/pulls?per_page=${DEFAULT_PER_PAGE}&page=${page}")

	# seperate the headers and body into 2 variables
	local head=true
	local header=
	local body=
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
	done < <(echo "${response}")

	get_last_page "${header}"

	local shas=$(echo $body | jq --raw-output '.[] | {id: .number, sha: .head.sha} | tostring')

	for s in $shas; do
		local id=$(echo $s | jq --raw-output '.id')
		local sha=$(echo $s | jq --raw-output '.sha')

		# send the request
		response=$(curl -sSL -H "${AUTH_HEADER}" -H "${API_HEADER}" "${URI}/repos/${repo}/commits/${sha}/status")

		# find the statuses that are not success
		local statuses=$(echo $response | jq --raw-output '. | select(.state != "success") | .statuses | .[] | select(.state != "success") | {context: .context, state: .state} | tostring')

		for status in $statuses; do
			local context=$(echo $status | jq --raw-output '.context')
			local state=$(echo $status | jq --raw-output '.state')

			if [[ "$context" != docker* ]]; then
				echo "Rebuilding pull request ${id} for context $context, build had state ${state}"
				data='{"repo":"'${repo}'","context":"'${context}'","number":'${id}'}'
				curl -ssL -X POST --user "${LEEROY_AUTH}" --data "${data}" ${LEEROY_URI}/build/custom
			fi
		done
	done
}

main(){
	local repo=$1
	: ${page:=1}

	if [[ -z "$repo" ]]; then
		echo "Pass a repo as the first arguement, ex. docker/docker"
		return 1
	fi

	rebuild_pulls "${repo}" "${page}"

	if [ ! -z "$LAST_PAGE" ] && [ "$LAST_PAGE" -ge "$page" ]; then
		for page in  $(seq $((page + 1)) 1 ${LAST_PAGE}); do
			echo "On page ${page} of ${LAST_PAGE}"
			rebuild_pulls "${repo}" "${page}"
		done
	fi
}

main "docker/docker"
