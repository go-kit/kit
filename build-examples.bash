#!/usr/bin/env bash

# This script builds all examples by looking for packages with "func main".

set -e

function go_files { grep -rl 'func main' examples ; }
function filter { grep -v -e gen-go ; }
function remove_relative_prefix { sed -e 's/^\.\///g' ; }

function directories {
	go_files | filter | remove_relative_prefix | while read f
	do
		dirname $f
	done
}

function unique_directories { directories | sort | uniq ; }

PATHS=${1:-$(unique_directories)}

function build {
	for path in $PATHS
	do
		echo building ./$path
		(cd ./$path && go build .)
	done
}

build

