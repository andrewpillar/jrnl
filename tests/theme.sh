#!/bin/sh

set -ex

DIR=$(mktemp -d)

theme_name="test-theme"

cd "$DIR"

jrnl init

printf "post layout\n" > _layouts/post
printf "page layout\n" > _layouts/page

mkdir _layouts/partials

printf "partial layout\n" > _layouts/partials/partial

jrnl gen-style solarized-light > _site/assets/style.css

jrnl theme save "$theme_name"

jrnl theme ls | grep "$theme_name"

rm -r _layouts/*
rm _site/assets/style.css

jrnl theme use "$theme_name"

[ -f _layouts/post ]
[ -f _layouts/page ]
[ -f _layouts/partials/partial ]
[ -f _site/assets/style.css ]

cd -

rm -rf "$DIR"
