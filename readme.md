# jrnl

jrnl is a simple static site generator written in Go, designed mainly for blogging. It takes posts written in Markdown, and converts them into HTML. The generate HTML files are then copied to a remote destination, this can either be a location on disk, or a remote HTTP server. Unlike most other static site generators jrnl does not serve the content that is generated. This article shall serve as a brief introduction to jrnl.

* [Installation](#installation)
* [Creating a Journal](#creating-a-journal)
* [Directory Structure](#directory-structure)
* [Posts](#posts)
* [Categories](#categories)
* [Post Indexing](#post-indexing)
* [Layouts](#layouts)
* [Themes](#themes)
* [Remotes](#remotes)
* [Publishing](#publishing)

## Installation {#installation}

If you have Go installed, then you can install jrnl with `go get`.

```
$ go get github.com/andrewpillar/jrnl
```

## Creating a Journal {#creating-a-journal}

Now that jrnl has been installed we can go ahead and create our first journal with the command `jrnl init`. This will create a journal in the current directory from which the command is run. However we can pass a directory name to the `init` command to initialise a journal inside that directory.

```
$ jrnl init blog
journal initialized, set the title with 'jrnl title'
```

This will create all the necessary files, and directories in the `blog` directory.

## Directory Structure {#directory-structure}

Below is the directory layout of a journal once initialised.

```
├── _layouts
├── _posts
├── _site
|   └── assets
├── _themes
└── _meta.yml
```

| Directory/File | Purpose                                                           |
|----------------|-------------------------------------------------------------------|
| `_layouts`     | Stores templates used for generating the site pages.              |
| `_posts`       | Stores all of the Markdown posts.                                 |
| `_site`        | Contains the final generated HTML that can be copied to a remote. |
| `_site/assets` | Assets for the site, such as CSS, and JS.                         |
| `_themes`      | Stores all available themes that can be used.                     |
| `_meta.yml`    | File containing meta-data about the journal.                      |

## Posts {#posts}

Posts in journal are created with the `jrnl post` command. This will take the title of the new post, and optionally a category for the post if one is passed using either category flag `-c` or `--category`. This will drop you in an editor as specified in the `_meta.yml` file. Every post that is created will have a block of YAML at the top known as front-matter. This block of YAML will contain the title of the post, the time it was created, and updated at, as well as the layout to use for the post when publishing, and whether or not it should be indexed.

```
$ jrnl post "Introducing jrnl"
```

```
---
title: Introducing jrnl
index: true
layout: post
createdAt: 2018-10-07T08:35
updatedAt: 2018-10-07T08:35
---
jrnl is a simple static site generator written in Go, designed mainly for blogging. It takes posts written in Markdown...
```

Posts can also be edited via the `jrnl edit` command. Running this command will updated the `updatedAt` field in the front-matter to the time of the edit. This command will take the ID of the post that you want to edit, to see which posts are available you can run `jrnl ls` to get a full list.

```
$ jrnl ls
introducing-jrnl
$ jrnl edit introducing-jrnl
```

## Categories {#categories}

As mentioned above every post created with journal can belong to a single category via use of the category flag.

```
$ jrnl post "Introducing jrnl" -c "Programming
```

This will store the newly created post file in a directory within `_posts` named for that new category.

```
├── _posts
|   └── programming
|       └── introducing-jrnl.md
```

Sub-categories can also be created just by using a `/` as the delimeter between each category.

```
$ jrnl post "Introducing jrnl" -c "Programming/Go"
```

## Post Indexing {#post-indexing}

Posts in journal can be indexed if they have the `index: true` property set in their front-matter. What this means is that during the publishing phase an `index.html` page will be created with a list of all the posts that have been created. Journal provides the ability to generate multiple indexes for posts. For example you will have a site-wide index which will list *all* posts created across the site, then you will have a category-wide index for all posts in that category, and another index for all posts created on a day, and so on.

All of these indexes are specified in the `_meta.yml` file, along with which layout file should be used for each index that is created.

## Layouts {#layouts}

Layouts are files that will define how parts of the site will look, whether it's a post or an index page. jrnl uses Go's `text/template` for templating, for more information on how to use them refer to the official [documentation](https://golang.org/pkg/text/template)

```html
<!DOCTYPE HTML>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <title>{{.Title}}</title>
    </head>
    <body>
        {{.Post.Body}}
    </body>
</html>
```

When a journal has been initialised the `_layouts` directory will be empty. It will be up to you to create the different layouts that you wish to use. However under the hood `jrnl` does expect some layout files to exist for indexing purposes. In the `_meta.yml` file there are the following group of properties:

```
indexlayouts:
  index: ""
  day: ""
  month: ""
  year: ""
  category: ""
  categoryday: ""
  categorymonth: ""
  categoryyear: ""
```

These define the layout files to be used by the different indexes on your site if you wish to have posts index. For example the `indexlayouts.categoryday` would be used for creating an `index.html` page underneath a category, on a given day. This would display all of the indexed posts made on that day, within that category.

## Themes {#themes}

Simply put themese are just tarballs of the site's `assets` directory, and of the `_layouts` directory. The current theme in use will be stored in the `_meta.yml`, and new themes can be created, saved, and deleted via the `jrnl theme` command.

You can check to see if a theme is in use just by running `jrnl theme`.

```
$ jrnl theme
no theme being used
$
```

Creating, or updating a theme is done by running `jrnl theme save`. If you currently don't have a theme loaded then the `save` command will expect an argument to be passed, which will be the name of the new theme. Otherwise the `save` command will overwrite the current theme in use.

Themes can be loaded by running `jrnl theme use`. This will take the tarball from the `_themes` directory, extract it, and replace the necessary directories.

All available themes can be listed with `jrnl theme ls`, and themes can be remove with `jrnl theme rm`.

## Remotes {#remotes}

As mentioned previousl jrnl does not serve the content it generates. Instead it simply copies everything in the `_site` directory to a remote. Remotes are managed via the `jrnl remote` command. To create a new remote simply run `jrnl remote set`.

```
$ jrnl remote set site me@andrewpillar.com: -d
```

Here we're setting up a new remote called `site`, that is pointing to the home directory on `andrewpillar.com`, and we're specifiying it as the default remote to use with the `-d` flag.

Now that this remote is setup our site will be copied to that server with SSH whenever we run `jrnl publish`.

When setting up a remote we can also specify which identity file to use for the SSH connection, along with the port to use, with the `-i`, and `-p` flags respectively.

```
$ jrnl remote set site me@andrewpillar.com: -d -p 1234 -i ~/.ssh/id_personal-site
```

## Publishing {#publishing}

Posts can be published with the `jrnl publish` command. This will take all of the posts that have been created, and produce the necessary indexes, if any are being used, and transform them into HTML using the layouts that are specified in each post.

Once this is done every file written to the `_sites` directory will be copied to the default remote if a default has been set. Different remotes can be used when publishing by passing the `-r` flag, followed by the name of the remote.

```
$ jrnl publish -r local
```

Drafts can also be published too by passing the `-d` flag. This will simply produce the indexes, and convert the Markdown to HTML, but it won't copy anything to the remote.

```
$ jrnl publish -d
```

My goal with this project was to learn some Go, and build a static site generator that is simple to use, and stays out of your way. I still need to go through the code base and add testing, but with regards to the functionality of the tool, I am happy with it.
