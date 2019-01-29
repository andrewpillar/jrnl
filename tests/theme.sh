#!/bin/sh

set -ex

DIR=$(mktemp -d)

theme_name="test-theme"

cd "$DIR"

jrnl init

printf "theme1: post layout\n" > _layouts/post

mkdir _layouts/partials

printf "theme1: partial layout\n" > _layouts/partials/partial

jrnl theme save theme1

printf "theme2: page layout\n" > _layouts/page
jrnl gen-style solarized-light > _site/assets/style.css

jrnl theme save theme2

jrnl theme use theme1

[ ! -f _site/assets/style.css ]
[ ! -f _layouts/page ]

[ -f _layouts/post ]
[ -f _layouts/partials/partial ]

jrnl theme use theme2

[ -f _layouts/page ]
[ -f _layouts/post ]
[ -f _layouts/partials/partial ]
[ -f _site/assets/style.css ]

cd -

rm -rf "$DIR"
