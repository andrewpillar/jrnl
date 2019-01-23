#!/bin/sh

set -ex

DIR=$(mktemp -d)

pushd "$DIR" > /dev/null

jrnl init > /dev/null

entries="First Post
Introducing jrnl"

IFS=$'\n'
while read -r e; do
	EDITOR=./tests/editor.sh jrnl post "$e"
done <<EOF
$entries
EOF

[ $(jrnl ls | wc -l) -eq 2 ]

slugs="first-post
introducing-jrnl"

while read -r s; do
	jrnl ls | grep "$s"
done <<EOF
$slugs
EOF

popd > /dev/null

rm -rf "$DIR"
