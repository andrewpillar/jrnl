#!/bin/sh

set -e

DIR="tests"

if [ ! -z "$1" ]; then
	printf "=== Running test script: %s\n" "$t"
	./"$DIR"/"$1"
	exit 0
fi

for t in $(ls "$DIR"); do
	printf "=== Running test script: %s\n" "$t"
	./"$DIR"/"$t"
done
