#!/bin/sh

set -e

front_matter="---
title: About
layout: \"\"
---\n"

title="About"
id="about"

pushd $(mktemp -d) > /dev/null

jrnl init > /dev/null
EDITOR=./tests/editor.sh jrnl page "$title" > /dev/null

if ! jrnl ls | grep -q "$id"; then
	printf "[FAILED] could not find '$id' with 'jrnl ls'\n"
	jrnl ls
	exit 1
fi

printf "[  OK  ] found '$id' with 'jrnl ls'\n"

expected=$(mktemp)
actual=$(mktemp)

printf "%b" "$front_matter" > $expected
head -n 4 _pages/"$id".md > $actual

if ! diff $expected $actual > /dev/null; then
	printf "[FAILED] front matter does not match expected\n"
	diff -u $expected $actual
	exit 1
fi

printf "[  OK  ] front matter matches expected\n"

popd > /dev/null
