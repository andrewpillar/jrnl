#!/bin/sh

set -ex

DIR=$(mktemp -d)

page_title="About"
page_id="about"

post_title="First Post"
post_id="first-post"

cat="TV Shows / The Leftovers"
cat_post_title="Penguin One, Us Zero"
cat_post_id="tv-shows/the-leftovers/penguin-one-us-zero"

pushd "$DIR" > /dev/null

jrnl init

EDITOR=./tests/editor.sh jrnl page "$page_title"
EDITOR=./tests/editor.sh jrnl post "$post_title"
EDITOR=./tests/editor.sh jrnl post -c "$cat" "$cat_post_title"

jrnl rm "$page_id" "$post_id" "$cat_post_id"

[ -z $(ls _pages) ]
[ -z $(ls _posts) ]

popd > /dev/null

rm -rf "$DIR"
