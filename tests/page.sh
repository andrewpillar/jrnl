#!/bin/sh

set -ex

DIR=$(mktemp -d)

front_matter="---
title: About
layout: \"\"
---\n"

title="About"
id="about"

pushd "$DIR" > /dev/null

jrnl init
EDITOR=./tests/editor.sh jrnl page "$title" > /dev/null

jrnl ls | grep "$id"

expected=$(mktemp)
actual=$(mktemp)

printf "%b" "$front_matter" > "$expected"
head -n 4 _pages/"$id".md > "$actual"

diff -u "$expected" "$actual"

popd > /dev/null

rm -rf "$DIR" "$expected" "$actual"
