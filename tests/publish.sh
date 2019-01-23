#!/bin/sh

set -ex

DIR=$(mktemp -d)

dir_now=$(date +"%Y/%m/%d")

remote=$(mktemp -d)

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

sed -i 's/layout: ""/layout: layout/g' _pages/"$page_id".md
sed -i 's/layout: ""/layout: layout/g' _posts/"$post_id".md
sed -i 's/layout: ""/layout: layout/g' _posts/"$cat_post_id".md

touch _layouts/layout

jrnl remote-set "$remote"
jrnl publish

[ -d "$remote/assets" ]
[ -f "$remote/$page_id/index.html" ]
[ -f "$remote/$dir_now/$post_id/index.html" ]
[ -f "$remote/$(dirname "$cat_post_id")/$dir_now/$(basename "$cat_post_id")/index.html" ]

popd > /dev/null

rm -rf "$DIR" "$remote"
