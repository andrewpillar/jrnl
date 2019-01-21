# jrnl

jrnl is a simple static site generator. It takes posts and pages written in Markdown, transforms them to HTML, and copies those HTML files over to a remote. Unlike most other static site generators jrnl does not serve the content it generates.

## Quick Start

If you have Go installed then you can simply clone this repository, and run `go install`, assuming you have `~/go/bin` added to your `$PATH`.

```
$ git clone https://github.com/andrewpillar/jrnl.git
$ cd jrnl
$ go install
```

Once installed you can create a new journal by running `jrnl init`.

```
$ jrnl init my-blog
journal initialized, set the title with 'jrnl title'
```

We can now change into the new directory and start creating posts with the `jrnl post` command.

```
$ cd my-blog
my-blog $ jrnl post "Introducing jrnl"
```

This will caused jrnl to drop you into a text editor, as specified via `$EDITOR` for editing the newly created post file. You will notice at the top of the post is a block of YAML with a handful of attributes. This is called front matter and contains meta-data about the post. For now let's set the `layout` attribute to `post`, and write up the post's content.

```
---
title: Introducing jrnl
index: false
layout: post
createdAt: 2006-01-02T15:04:05
updatedAt: 2006-01-02T15:04:05
---
jrnl is a simple static site generator. It takes posts and pages written in Markdown...
```

Now that we have written up our post the next thing we need to do is create a layout for our post. Above we specified the layout for our post via the `layout` property in the front matter. Now we need to create that layout file for jrnl to use during transformation. All layout files in jrnl are stored in the `_layouts` directory. So let's create a `post` layout file, here's what one could look like:

```html
<!DOCTYPE HTML>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <title>{{.Site.Title}}</title>
    </head>
    <body>
        <h1>{{.Post.Title}}</h1>
        <div>{{.Post.Body}}</div>
    </body>
</html>
```

Nothing too special here, just some basic HTML. Layout files in jrnl utilise Go's [text/template](https://golang.org/pkg/text/template) library for templating. As you can see we are placing the post's title and body into the HMTL. For more information about the different layouts used by jrnl, and the variables within these layouts read the [Layouts]() section of this readme.

So we have a post, and a layout file. Next let's set a title for our journal, we were after all using the `.Site.Title` variable in our post layout. We can do this by running `jrnl title`.

```
$ jrnl title "My Blog"
```

Now, before we can publish our journal we need to set a remote. A remote can either be a local filesystem path, or a remote SCP URL for copying the HTML files to. To keep things simple lets just use a local filesystem path for now.

```
$ jrnl remote-set /tmp/blog-remote
$ jrnl publish
```

If we perform an `ls` on the `/tmp/blog-remote` remote that we created we should see the directory populated with some files generated from jrnl.

```
$ ls /tmp/blog-remote
2006 assets
```

And there is our site. Not much right now, just an empty `assets` directory and a path pointing to our published post which we can now view in a browser of our choice.

```
$ firefox 2006/01/02/introducing-jrnl/index.html
```


