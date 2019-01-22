# jrnl

jrnl is a simple static site generator. It takes posts and pages written in Markdown, transforms them to HTML, and copies those HTML files over to a remote. Unlike most other static site generators jrnl does not serve the content it generates.

* [Quick Start](#quick-start)
* [Initializing a Journal](#initializing-a-journal)
* [Directory Structure](#directory-structure)
* [Pages and Posts](#pages-and-posts)
* [Categories](#categories)
* [Front Matter](#front-matter)
* [Indexing](#indexing)
* [Themes](#themes)
* [Remote](#remote)

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

Nothing too special here, just some basic HTML. Layout files in jrnl utilise Go's [text/template](https://golang.org/pkg/text/template) library for templating. As you can see we are placing the post's title and body into the HMTL.

So we have a post and a layout file. Next let's set a title for our journal, we were after all using the `.Site.Title` variable in our post layout. We can do this by running `jrnl title`.

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

## Initializing a Journal

A new journal can be initialized by running `jrnl init`. If this command is provided an argument then the journal will be initialized in the given directory. Otherwise the journal initialized in the current directory.

```
$ jrnl init my-blog
journal initialized, set the title with 'jrnl title'
```

## Directory Structure

Below is the directory structure of a journal once initialized.

```
├── _layouts
├── _pages
├── _posts
├── _site
|   └── assets
├── _themes
└── config
```

| Directory / File | Purpose |
|------------------|---------|
| `_layouts` | Stores templates used for generating posts and pages. |
| `_pages` | Where the Markdown pages are stored. |
| `_posts` | Where the Markdown posts are stored. |
| `_site` | Contains the final generated HTML files. |
| `_site/assets` | Assts for the site, such as CSS, JS, and any images. |
| `_themes` | Where all the journal themes are stored. |
| `config` | Configuration file for the journal. |

## Pages and Posts

jrnl has two types of content, pages and posts. A page is a simple content type, it is made up of a title, a layout, and a body. Pages are created with the `jrnl page` command. This takes the title of the page to be created as its only argument.

```
$ jrnl page About
```

Posts are similar to pages in how they are created. The main difference is that posts can be categorized, and contain meta-data about when that post was created, and updated.

```
$ jrnl post "Introducing jrnl"
```

Whenever a page or post is created, or edited jrnl will open up the source Markdown file in the editor that you havbe set via the `$EDITOR` environment variable. Upon creation of a new page or post the source Markdown file will be pre-populated with some [front matter](#front-matter) which will store some meta-data about the page or post.

Each time a page or post is created its ID will be written to `stdout`. All of the posts and pages that have been created can be viewed with the `jrnl ls` command.

```
$ jrnl ls
about
introducing-jrnl
```

These IDs can then be passed to either `jrnl edit` or `jrnl rm` for modification or removal respectively.

## Categories

jrnl allows for posts to be stored in categories, and sub-categories. To add a post to a category simply pass the `-c` flag to the `jrnl post` command.

```
$ jrnl post "Pengun One, Us Zero" -c "TV Shows / The Leftovers"
```

This will store the post Markdown file in a sub-directory beneath the `_posts` directory.

```
├── _posts
|   └── tv-shows
|       └── the-leftovers
|           └── penguin-one-us-zero.md
```

Under the hood categories are nothing more than additional directories to store posts in.

## Front Matter

Front matter is a block of YAML that sits at the top of each page or post in the Markdown source file. It contains meta-data about the piece of content being modified. The front matter can vary between pages and posts in regards to what it stores, but both types of content will have a `title` property, and a `layout` property.

Below is some example front matter for a post.

```
---
title: Penguin One, Us Zero
index: true
layout: post
createdAt: 2006-01-02T15:04:05
updatedAt: 2006-01-02T15:04:05
---
```

| Property | Purpose | Page or Post |
|----------|---------|--------------|
| `title` | The title of the post or page.| Both |
| `index` | Whether or not to add the posts to any indexes in the journal. | Post |
| `layout` | The name of the layout file to use from the `_layouts` directory for templating. | Both |
| `createdAt` | When the post was created. | Post |
| `updatedAt` | When the post was last edited. | Post |

## Layouts

Layouts are text files that define how a page or a post will look once published. jrnl uses Go's [text/template](https://golang.org/pkg/text/template) library for templating. When a journal has been initialized the `_layouts` directory will be empty, it will be up to you to create the necessary layout files. Below is an example layout file for a post:

```html
<!DOCTYPE HTML>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <title>{{.Site.Title}}</title>
    </head>
    <body>{{.Post.Body}}</body>
</html>
```

All layout files will have access to the `.Site` value. This value allows for retrieving the title of the journal, all of the pages in the journal, and all of the categories.

```html
<ul>
    {{range $i, $p := .Site.Pages}}
        <li><a href="{{$p.Href}}">{{$p.Title}}</a></li>
    {{end}}
</ul>
```

```html
<ul>
    {{range $i, $c := .Site.Categories}}
        <li><a href="{{$c.Href}}">{{$c.Name}}</a></li>
    {{end}}
</ul>
```

The layout used by a page will be passed the `.Page` value, and the layout used by a post will be passed the `.Post` value.

You will also have access to the `partial` function within a layout file. This can allow you to create re-usable parts of a layout. The partial function takes the name of the layout file to include, followed by the data to pass through.

```html
{{partial "categories" .}}
```

In the above example we pull in the `categories` partial, and pass it `.`, which means to pass through all the data in the current template.

## Indexing

jrnl allows for `index.html` files to be generated that simply contain a list of the posts created. There are four types of indexes in journal: `day`, `month`, `year`, and `all`.

The `day`, `month`, and `year` indexes will list all of the posts that were made during the respective time period. Whereas the `all` index will list all of the posts that were ever made.

To create a special index layout simply create an `_index` directory where you want the index to reside. And populate that directory with the layout files for the types of index you want.

For example, to create a daily index for the `TV Shows` category you would create the `_index` directory beneath the `_posts/tv-shows` directory, and create a layout filed called `day`.

```
$ mkdir _posts/tv-shows/_index
$ vim _posts/tv-shows/_index/day
```

Below is what the layout for the daily index might look like:

```html
<!DOCTYPE HTML>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <title>{{.Site.Title}}</title>
    </head>
    <body>
        <h1>{{.Category.Name}} - Posts from {{.Time.Format "Mon 02 Jan 2006"}}</h1>
        {{range $i, $p := .Posts}}
            <div>{{$p.Title}} - {{$p.Preview}}</div>
        {{end}}
    </body>
</html>
```

Indexes which are time based will be passed the `.Time` value which is a Go `time.Time` struct. This will only be as accurate as the index for which it is used.

Any indexes which are for a category will also be passed the `.Category` value for the category of the index itself.

To have a post indexed simply edit its front matter, and set the `index` property to `true`.

## Themes

Themes in jrnl are just a tarball of the `_layouts` directory, and the `_site/assets` directory. The current theme is stored in the `config` file. All theme management in jrnl is done via the `jrnl theme` command. You can check to see if a theme is in use by running `jrnl theme`.

```
$ jrnl theme
no theme being used
```

Creating, or saving a theme is done by running `jrnl theme save`. If no argument is given then it will overwrite the current them in use, otherwise it will create a new theme with the given name.

Themes can be used by running the `jrnl theme use` command, and passing it the name of theme you wish to use.

All available themes can be listed with `jrnl theme ls`, and themes can be deleted with `jrnl theme rm`.

>**Note:** Right now there is a limitation surrounding themes, and index layouts. The index layouts will currently be ignored during theme creation, as they are intrinsically tied to the posts that have been created. Right now if you wish to have an index layout be part of a theme the solution is to store the index layouts in the `_layouts` directory, and call `partial` in the main index layout file to include the layout itself.

## Remote

Each journal has a remote. A remote is where the contents of the `_site` directory is copied to, this could either be a local filesystem path, or an SCP URL. The remote can be set via the `jrnl remote-set` command, and is stored in the `config` file.

```
$ jrnl remote-set me@andrewpillar.com:/var/www/andrewpillar.com
```

## Publishing

To publish a journal simply run `jrnl publish`. If a remote has been set the contents of the  `_site` directory  will be copied across. If the remote is an SCP URL, then jrnl will attempt an SSH connection on port `22` to the remote, and use the `~/.ssh/id_rsa` key for authentication.

```
$ jrnl publish
```

Drafts can be published by passing the `-d` flag to the `jrnl publish` command. This will only produce the HTML files instead of copying them over.

```
$ jrnl publish -d
```
