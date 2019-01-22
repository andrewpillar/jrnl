#!/bin/sh

set -e

front_matter="---
title: About
layout: \"\"
---\n"

pushd $(mktemp -d) > /dev/null

jrnl init > /dev/null
EDITOR=./tests/editor.sh jrnl page About > /dev/null

if [ -f _pages/about.md ]; then
	printf "[  OK  ] page 'About' created\n"
else
	printf "[FAILED] page 'About' not created\n"
	exit 1
fi

expected=$(mktemp)
actual=$(mktemp)

printf "%b" "$front_matter" > $expected
head -n 4 _pages/about.md > $actual

if ! diff $expected $actual > /dev/null; then
	printf "[FAILED] front matter does not match expected\n"
	diff -u $expected $actual
	exit 1
fi

printf "[  OK  ] front matter matches expected\n"

popd > /dev/null
