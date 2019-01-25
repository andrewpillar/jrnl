#!/bin/sh

DIR=$(mktemp -d)

remote=$(mktemp -d)

cd "$DIR"

jrnl init
jrnl remote-set "$remote"

grep "$remote" jrnl.yml

cd -

rm -rf "$DIR" "$remote"
