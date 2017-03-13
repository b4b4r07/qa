package main

const (
	SCRIPT_BRANCHES string = `
	for repo in /var/www/vhosts/*
	do
		cd "$repo/current" &>/dev/null
		branch="$(git rev-parse --abbrev-ref HEAD)"
		echo "$repo $branch"
	done
	`
)
