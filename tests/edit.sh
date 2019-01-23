#!/bin/sh

set -ex

DIR=$(mktemp -d)

iso_now=$(date +"%Y-%m-%dT%H:%M")

title="First Post"
id="first-post"

cd "$DIR"

jrnl init
EDITOR=./tests/editor.sh jrnl post "$title"

sed -i "s/updatedAt: $iso_now/updatedAt: 2006-01-02T15:04/g" _posts/"$id".md

EDITOR=./tests/editor.sh jrnl edit "$id"

grep "$iso_now" _posts/"$id".md

cd -

rm -rf "$DIR"
