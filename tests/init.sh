#!/bin/sh

set -e

jrnl_files="config"

jrnl_dirs="_layouts
_pages
_posts
_site
_themes"

pushd $(mktemp -d) > /dev/null

jrnl init

while read -r f; do
	if [ -f $f ]; then
		printf "[  OK  ] %s exists\n" $f
	else
		printf "[FAILED] %s does not exist\n" $f
		exit 1
	fi
done <<EOF
$jrnl_files
EOF

while read -r d; do
	if [ -d $d ]; then
		printf "[  OK  ] %s exists\n" $d
	else
		printf "[FAILED] %s does not exist\n" $d
		exit 1
	fi
done <<EOF
$jrnl_dirs
EOF

popd > /dev/null

pushd $(mktemp -d) > /dev/null

jrnl init blog

while read -r f; do
	if [ -f blog/$f ]; then
		printf "[  OK  ] blog/%s exists\n" $f
	else
		printf "[FAILED] blog/%s does not exist\n" $f
		exit 1
	fi
done <<EOF
$jrnl_files
EOF

while read -r d; do
	if [ -d blog/$d ]; then
		printf "[  OK  ] blog/%s exists\n" $d
	else
		printf "[FAILED] blog/%s does not exist\n" $d
		exit 1
	fi
done <<EOF
$jrnl_dirs
EOF

popd > /dev/null
