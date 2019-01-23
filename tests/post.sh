#!/bin/sh

set -ex

DIR=$(mktemp -d)

iso_now=$(date +"%Y-%m-%dT%H:%M")

front_matter="---
title: First Post
index: false
layout: \"\"
createdAt: $iso_now
updatedAt: $iso_now
---\n"

title="First Post"
id="first-post"

cat_title="Penguin One, Us Zero"
cat_id="tv-shows/the-leftovers/penguin-one-us-zero"
cat="TV Shows / The Leftovers"

pushd "$DIR" > /dev/null

jrnl init
EDITOR=./tests/editor.sh jrnl post "$title"

jrnl ls | grep "$id"
touch _posts/"$id".md

expected=$(mktemp)
actual=$(mktemp)

printf "%b" "$front_matter" > "$expected"
head -n 7 _posts/first-post.md > "$actual"

diff -u "$expected" "$actual"

EDITOR=./tests/editor.sh jrnl post -c "$cat" "$cat_title"

jrnl ls | grep "$cat_id"
touch _posts/"$cat_id".md

popd > /dev/null

rm -rf "$DIR" "$expected" "$actual"
