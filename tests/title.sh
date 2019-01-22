#!/bin/sh

set -e

title="My Blog"

pushd $(mktemp -d) > /dev/null

jrnl init > /dev/null
jrnl title "$title"

if ! grep -q "$title" config; then
	printf "[FAILED] journal title not set\n"
	cat config
	exit 1
fi

printf "[  OK  ] journal title set\n"

popd > /dev/null
