#!/bin/sh

set -ex

DIR=$(mktemp -d)

title="My Blog"

cd "$DIR"

jrnl init
jrnl title "$title"

grep "$title" jrnl.yml

cd -

rm -rf "$DIR"
