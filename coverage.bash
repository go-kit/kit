#!/usr/bin/env bash

# This script runs the cover tool on all packages with test files. If you set a
# WEB environment variable, it will additionally open the web-based coverage
# visualizer for each package.

function go_files { find . -name '*_test.go' ; }
function filter { grep -v '/_' ; }
function remove_relative_prefix { sed -e 's/^\.\///g' ; }

function directories {
	go_files | filter | remove_relative_prefix | while read f
	do
		dirname $f
	done
}

function unique_directories { directories | sort | uniq ; }

PATHS=${1:-$(unique_directories)}

function package_names {
	for d in $PATHS
	do
		echo github.com/go-kit/kit/$d
	done
}

function report {
	package_names | while read pkg
	do
		go test -coverprofile=cover.out $pkg
		if [ -n "${WEB+x}" ]
		then
			go tool cover -html=cover.out
		fi
	done
	rm cover.out
}

report

