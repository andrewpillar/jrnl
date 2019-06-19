package util

import (
	"regexp"
	"strings"
)

var (
	reslug = regexp.MustCompile("[^a-zA-Z0-9]")
	redup  = regexp.MustCompile("-{2,}")
)

func Slug(s string) string {
	s = strings.TrimSpace(s)

	s = reslug.ReplaceAllString(s, "-")
	s = redup.ReplaceAllString(s, "-")

	return strings.ToLower(s)
}
