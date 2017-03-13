package main

const (
	// should be LTSV
	SCRIPT_BRANCHES string = `
	for repo in /var/www/vhosts/*
	do
		cd "$repo/current" &>/dev/null
		branch="$(git rev-parse --abbrev-ref HEAD)"
		printf "$repo\t$branch\n"
	done
	`
)
