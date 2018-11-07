---
title: Post
index: true
layout: post
createdAt: 2006-01-02T15:04
updatedAt: 2006-01-02T15:04
---
Files in the `testdata` directory are used for unit testing jrnl. These files will serve to partially emulate a jrnl directory, and hopefully demonstrate how posts, pages, and categories work in jrnl. Files with the `.golden` suffix are used for comparing complex strings that are output during the tests. This saves having to place this data inside the test code itself.

Another paragraph here is used to pad out this test file a bit more. When a post is loaded the front matter will be stripped, and the first line after the front matter will be read in as the post's preview. Everything else after the front matter will be read into the post's body.

Front matter on a post is a block of YAML at the beginning of the file that contains meta data about the post, such as the title, layout, whether or not it should be indexed, and the post's timestamps.
