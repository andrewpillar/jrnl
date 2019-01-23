#!/bin/sh

set -ex

DIR=$(mktemp -d)

title="My Blog"

pushd "$DIR" > /dev/null

jrnl init
jrnl title "$title"

grep "$title" config

popd > /dev/null

rm -rf "$DIR"
