#!/bin/sh

DIR=$(mktemp -d)

remote=$(mktemp -d)

pushd "$DIR" > /dev/null

jrnl init
jrnl remote-set "$remote"

grep "$remote" config

popd > /dev/null

rm -rf "$DIR" "$remote"
