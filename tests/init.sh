#!/bin/sh

set -ex

DIR=$(mktemp -d)

jrnl_files="config"
jrnl_dirs="_layouts
_pages
_posts
_site
_themes"

cd "$DIR"

jrnl init

while read -r f; do
	[ -f "$f" ]
done <<EOF
$jrnl_files
EOF

new_dir="blog"

jrnl init "$new_dir"

while read -r d; do
	[ -d "$new_dir/$d" ]
done <<EOF
$jrnl_dirs
EOF

cd -

rm -r "$DIR"
