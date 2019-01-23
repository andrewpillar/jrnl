#!/bin/sh

set -e

#iso_now=$(date +"%Y-%m-%dT%H:%M")
#
#front_matter="---
#title: First Post
#index: false
#layout: \"\"
#createdAt: $iso_now
#updatedAt: $iso_now
#---\n"
#
#pushd $(mktemp -d) > /dev/null
#
#jrnl init > /dev/null
#EDITOR=./tests/editor.sh jrnl post "First Post" > /dev/null
#
#if [ -f _posts/first-post.md ]; then
#	printf "[  OK  ] post 'First Post' created\n"
#else
#	printf "[FAILED] post 'First Post' not created\n"
#	exit 1
#fi
#
#expected=$(mktemp)
#actual=$(mktemp)
#
#printf "%b" "$front_matter" > $expected
#head -n 7 _posts/first-post.md > $actual
#
#if ! diff $expected $actual > /dev/null; then
#	printf "[FAILED] front matter does not match expected\n"
#	diff -u $expected $actual
#	exit 1
#fi
#
#printf "[  OK  ] front matter matches expected\n"
#
#popd > /dev/null
