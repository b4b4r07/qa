package main

const (
	// should be LTSV
	SCRIPT_BRANCHES string = `
	for path in /var/www/vhosts/*
	do
		(
		cd "$path/current" &>/dev/null
		name=$(basename $path)
		branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null)"
		date=$(git show --quiet --pretty=format:"%ar (%h)" "$branch")
		printf "name:$name\tpath:$path\tbranch:$branch\tdate:$date\n"
		) &
	done
	wait
	`
)
