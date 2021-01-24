# jrnl

jrnl is a simple static site generator. It takes posts and pages written in
Markdown, transforms them to HTML, and copies those HTML files over to a
remote. Unlike most other static site generators jrnl does not serve the
content it generates.

* [Quick start](#quick-start)
* [Initializing jrnl](#initializing-jrnl)
* [Directory structure](#directory-structure)
* [Pages and posts](#pages-and-posts)
* [Categories](#categories)
* [Front matter](#front-matter)
* [Layouts](#layouts)
* [Indexing](#indexing)
* [Themes](#themes)
* [Remote](#remote)
* [Publishing](#publishing)
* [Atom and RSS feeds](#atom-and-rss-feeds)

## Quick start

First clone the repository and run the `make.sh` script to build jrnl.

    $ git clone https://github.com/andrewpillar/jrnl
    $ cd jrnl && ./make.sh

This will produced a `jrnl` binary in the `bin` directory. Copy this file into
your `$PATH`,

    $ cp bin/jrnl ~/go/bin

Once installed you can create a new jrnl by running `jrnl init`.

    $ jrnl init my-blog

this will initialize everything you need within the `my-blog` directory. If you
do not pass an argument to this command then the jrnl will be initialized in the
current directory.

We can now change into the new directory and start creating posts with the
`jrnl post` command.

    $ jrnl post "Introducing jrnl"

this will cause jrnl to drop you into a text editor, as specified via `$EDITOR`
for editing the newly created post. You will notice at the top of the post is a
block of YAML with a handful of attributes. This is called front matter, and
contains meta-data about the post. For now let's set the `layout` attribute to
`post`, and write up the post's content.

    ---
    title: Introducing jrnl
    layout: post
    createdAt: 2001-01-02T15:04
    updatedAt: 2001-01-02T15:04
    ---
    jrnl is a simple static site generator.

Now that we have written up our post the next thing we need to do is created
a layout for our post. Above we specified the layout for our post via the
`layout` property in the front matter. Now we need to create that layout file
for jrnl to use during publishing. All layout files in jrnl are stored in the
`_layouts` directory. So let's create a layout file, here's what one could look
like:

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

Nothing too specield here, just some basic HTML. Layout files in jrnl utilize
Go's [text/template](https://golang.org/pkg/text/template) library for
templating. As you can see we are placing the post's title and body into the
HTML.

So we have a post and a layout file. Next let's set a title for our jrnl. We
can do this by running `jrnl config` and by specifying the `site.title` key.

    $ jrnl config site.title "My Blog"

Now, before we can publish our jrnl we need to set a remot. A remote can either
be a local filesystem path, or a remote SCP URL for copying the HTML files to.
To keep things simple lets just use a local filesystem path for now.

    $ jrnl config site.remote /tmp/blog-remote
    $ jrnl publish

If we peform an `ls` on the `/tmp/blog-remote` path we specified, we should see
the directory populated with some files generated from jrnl.

    $ ls /tmp/blog-remote
    2006 assets

And there is our site. Not much right now, just and empty `assets` directory and
a path pointing to our published post which we can now view in a browser.

    $ firefox 2006/01/02/introducing-jrnl/index.html

## Initializing jrnl

jrnl can be initialized by running `jrnl init`. If this command is provided an
argument then jrnl will be initialized in the given directory. Otherwise jrnl is
initialized in the current directory.

    $ jrnl init my-blog

## Directory structure

A jrnl is structured like so,

    ├── _data
    ├── _layouts
    ├── _pages
    ├── _posts
    ├── _site
    |   └── assets
    ├── _themes
    └── jrnl.toml

* `_data` - Stores binary data about the pages and posts.
* `_layouts` - Stores templates used for generating pages and posts.
* `_pages` - Stores the Markdown files for the pages.
* `_posts` - Stores the Markdown files for the posts.
* `_site` - Stores the final generated HTML files.
* `_site/assets` - Stores the static assets for the site, such as CSS, JS, and
any images.
* `_themes` - Stores the jrnl themes.
* `jrnl.toml` - The configuration file for jrnl.

## Pages and posts

jrnl has two types of content, pages and posts. A page is a simple content
type, it is made up of a title, a layout, and a body. Pages are created with
the `jrnl page` comman. This takes the title of the page to be created as its
only argument.

    $ jrnl page About

Posts are similar to pages in how they are created. The main different is that
posts can be categorized, and contain meta-data about when that post was
created, and updated.

    $ jrnl post "Introducing jrnl"

Whenever a page or post is created, or edited jrnl will open up the source
Markdown f ile in the editor that you have set via the `$EDITOR` environment
variable. Upon creation of a new page or post the source Markdown file will be
pre-populated with some [front matter](#front-matter) whcih will store some
meta-data about the page or post.

All of the posts and pages that have been created can be viewed with the
`jrnl ls` command.

    $ jrnl ls
    about
    introducing-jrnl

These IDs can be passed to `jrnl edit` or `jrnl rm` for modification or removal
respectively.

## Categories

jrnl allows for posts to be stored in categories, and sub-categories. To add a
post to a category simply pass the `-c` flag to the `jrnl post` command,

    $ jrnl post -c "TV Shows / The Leftovers" "Penguin One, Us Zero"

this will store the Markdown file in a sub-directory beneath the `_posts`
directory.

    ├── _posts
    |   └── tv-shows
    |       └── the-leftovers
    |           └── penguin-one-us-zero.md

Under the hood categories are nothing more than additional directories to store
posts in.

## Front matter

Front matter is a block of YAML that sits at the top of each page or post in the
Markdown source file. It contains meta-data about the pieve of content being
modified. The front matter can vary between pages and posts in regards to what
it strores, but both types of content will have a `title` and `layout` property.

* `title` (both) - The title of the page or post.
* `layout` (both) - The layout of the page or post.
* `createdAt` (post) - The time the post was created.
* `updatedAt` (post) - The time the post was updated.

## Layouts

Layouts are text files that define how a page or post will look once published.
jrnl uses Go's [text/template](https://golang.org/pkg/text/template) library for
templating. When jrnl has been initialized the `_layouts` directory will be
empty, it will be up to you to create the necessary layout files. Below is an
example layout file for a post:

    <!DOCTYPE HTML>
    <html lang="en">
        <head>
            <meta charset="utf-8">
            <title>{{.Site.Title}}</title>
        </head>
        <body>{{.Post.Body}}</body>
    </html>

All layout files will have access to the `.Site` value. This value allows for
retrieving the title, authorship, categories, and pages in the jrnl```html

    <ul>
        {{range $i, $p := .Site.Pages}}
            <li><a href="{{$p.Href}}">{{$p.Title}}</a></li>
        {{end}}
    </ul>
    
    <ul>
        {{range $i, $c := .Site.Categories}}
            <li><a href="{{$c.Href}}">{{$c.Name}}</a></li>
        {{end}}
    </ul>

The layout used by a page will be passed the `.Page` value, and the layout used
by a post will be passed the `.Post` value.

You will also have access to the `partial` function within a layout file. This
allows you to create re-usable parts of a layout. The partial function takes
the name of the layout file to include followed by the data to passthrough.

    {{partial "categories" .Site.Categories}}

## Indexing

jrnl can generate an `index.html` file at the root of the `_site` directory
that will list all of the posts created. One can also be created for each
category that will list each of the category's posts.

To have these index files created, simply specify an index layout file to use
in the `_layouts` directory. Use `index` for the `_site/index.html` file, and
`category-index` for a category specific `index.html` file.

## Themes

Themes in jrnl are just a tarball of the `_layouts` directory, and the
`_site/assets` directory. The current theme being used is stored in the
`jrnl.toml` configuration file. You can check to see if a theme is in use by
running `jrnl theme`.

    $ jrnl theme

Creating or saving a theme is done with `jrnl theme save`. If no argument is
given then it will overwrite the current theme in use, otherwise it will create
a new theme with the given name.

Themes can beu sed by running the `jrnl theme use` command, and passing it the
name of the theme you wish to use.

All available themes can be listed with `jrnl theme ls`, and themes can be
deleted with `jrnl theme rm`.

## Remote

Each jrnl has a remote. A remote is where the contents of the `_site` directory
is copied to. This can either be a location on disk, or an SCP URL. The remote
can be set via `jrnl config site.remote` and is stored in the `jrnl.toml` file.

    $ jrnl config site.remote me@andrewpillar.com:/var/www/andrewpillar.com

## Publishing

To publish a jrnl simply run `jrnl publish`. This will transform all of the
modified Markdown pages and posts into HTML, and generate the necessary
`index.html` files before copying them to the remote.

    $ jrnl publish

Drafts can be published by setting the `-d` flag. This will only produce the
HTML files instead of copying them over.

Each page and post that is published will be written to the `_data/hash` file.
This is used to determine which pages and posts should be copied to the remote
based on whether they have been modified.

## Atom and RSS feeds

Atom and RSS feeds can be generated by passing the `-a` and `-r` flags to the
`jrnl publish` command. Each of these flags will take a path to the file where
you would like the feed to be written,

    $ jrnl publish -a _site/atom.xml -r _site/rss.xml
