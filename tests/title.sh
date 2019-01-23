#!/bin/sh

set -ex

DIR=$(mktemp -d)

title="My Blog"

cd "$DIR"

jrnl init
jrnl title "$title"

grep "$title" config

cd -

rm -rf "$DIR"
