# jrnl

jrnl is a simple static site generator written in Go. It takes posts written in Markdown, and converts them to HTML. The generated HTML files are then copied to a remote destination, this can either be a location on disk, or to a remote server. Unlike most other static site generators jrnl does not serve the content that is generated. jrnl has support for static pages, posts, categories, sub-categories, themes, and post indexing.

* [Creating a Journal](#creating-a-journal)
* [Directory Structue](#directory-structure)
* [Pages and Posts](#pages-and-posts)
* [Categories](#categories)
* [Front Matter](#front-matter)
* [Post Indexing](#post-indexing)
* [Layouts](#layouts)
* [Themes](#themes)
* [Remotes](#remotes)
* [Publishing](#publishing)

## Creating a Journal

A new journal can be created by running `jrnl init`. If this command is provided an argument that the journal will be created in the given directory, otherwise the journal will be created in the current directory.

```
$ jrnl init blog
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
└── _meta.yml
```

| Directory/File | Purpose |
|----------------|---------|
| `_layouts` | Stores templates used for generating pages, and posts. |
| `_pages` | Stores the Markdown pages. |
| `_posts` | Stoes the Markdown posts. |
| `_site` | Contains the generated HTML files. |
| `_site/assets` | Assets for the site such as CSS, JS, and any images. |
| `_themes` | Stores all available themes that can be used. |
| `_meta.yml` | File containing meta-data about the journal. |

## Pages and Posts

jrnl has two types of content, pages, and posts. A page is a simple content type, it is made up of a title, a layout, and the body of content itself. An example of a page would be an About page, or a Contact page.

Pages are created via the `jrnl page` command.

```
blog $ jrnl page About
```

This will drop you into an editor with some pre-populated front matter for the page.

```
---
title: About
layout: page
---
This is the about page...
```

Pages can be edited with `jrnl edit`, and removed with `jrnl rm`. Each of these commands takes the page ID of the page you would want to edit or remove.

Posts are similar to pages in how they are created, and managed. You can create a new post with the `jrnl post` command.

```
blog $ jrnl post "Penguin One, Us Zero" -c "TV Shows/The Leftovers"
```

The `jrnl post` command will drop you into an editor with some pre-populated front matter, similar to the `jrnl page` command.

```
---
title: Penguin One, Us Zero
index: true
layout: post
createdAt: 2018-11-10T22:55
updatedAt: 2018-11-10T22:55
---
Season 1, episode 2 of The Leftovers...
```

The front matter for a post contains more information then the page, such as whether or not it should be indexed, and the timestamps for when the post was created, and last updated.

All of the pages, and posts created can be viewed with the `jrnl ls` command.

```
blog $ jrnl ls
about
tv-shows/the-leftovers/penguin-one-us-zero
```

This will print out all of the IDs of the pages, and posts, which can then be used with the `jrnl edit`, or `jrnl rm commands`.

```
blog $ jrnl rm about
blog $ jrnl ls
tv-shows/the-leftovers/penguin-one-us-zero
```

## Categories

Every post created with jrnl can belong to a category. This is done by setting the `-c` flag on the command `jrnl post` when creating a new post.

```
blog $ jrnl post "Penguin One, Us Zero" -c "TV Shows/The Leftovers"
```

This will store the new post under that category, if the given category does not exist then it will be created. Under the hood a category is nothing more than a directory.

```
├── _posts
|   └── tv-shows
|       └── the-leftovers
|           └── penguin-one-us-zero.md
```

jrnl has support for sub-categories via the `/` delimiter. When specifying the category name just use `/` to delimit the sub-categories.

Of course you can have standalone posts too which do not belong to a category, in this case your can drop the `-c` flag altogether.

```
blog $ jrnl post "Update - Nov 18"
```

```
├── _posts
|   └── tv-shows
|       └── the-leftovers
|           └── penguin-one-us-zero.md
└── update-nov-18.md
```

## Front Matter

jrnl uses front matter for both pages, and posts to store meta-data about the content. Below is an example of some front matter for a post.

```
---
title: Penguin One, Us Zero
index: true
layout: post
createdAt: 2018-11-10T22:55
updatedAt: 2018-11-10T22:55
---
```

| Property | Purpose |
|----------|---------|
| `title` | The title of the post, this will be whatever was given to `jrnl post`. |
| `index` | Whether or not the post should be indexed during publishing. |
| `layout` | The name of the file to use from the `_layouts` directory during publishing. |
| `createdAt` | When the post was created. |
| `updatedAt` | When the post was last updated, this will be set when `jrnl edit` is run. |

The front matter for a page is far simpler than that for a post, and only contains the `title`, and `layout` properties.

## Post Indexing

When a journal is published, `index.html` files will be created which display all of the posts that have been written in the journal. The `index` property in a post's front matter will tell jrnl whether or not to include that post in an index.

Below is what what the `_site` directory would look like once a journal has been published with indexes.

```
├── _site
|   ├── assets
|   ├── tv-shows
|   |   ├── the-leftovers
|   |   |   ├── 2018
|   |   |   |   ├── 11
|   |   |   |   |   ├── 10
|   |   |   |   |   |   ├── penguin-one-us-zero
|   |   |   |   |   |   └── index.html
|   |   |   |   |   └── index.html
|   |   |   |   └── index.html
|   |   |   └── index.html
|   |   └── index.html
|   ├── 2018
|   |   ├── 11
|   |   |   ├── 11 
|   |   |   |   ├── update-nov-18
|   |   |   |   └── index.html
|   |   |   └── index.html
|   |   └── index.html
|   └── index.html
```

As you can see multiple index files are created for each category, year, month, and day that a post was written on, along with a single site wide index that would list all of the posts that have been written. These indexes will only be created however if the corresponding layout file for that index has been specified in the `_meta.yml` file, as shown below.

```yaml
indexLayouts:
  index: indexes/index
  day: indexes/day
  month: indexes/month
  year: indexes/year
  category:
    index: indexes/category/index
    day: indexes/category/day
    month: indexes/category/month
    year: indexes/category/year
```

jrnl will search the `_layouts` directory for the above listed files to use as templates when it writes the index files. If no files can be found then jrnl will write the errors to `stderr`, but will continue with publishing the rest of the journal.

## Layouts

Layouts are files that define how pages, and posts will look once published. jrnl uses Go's `text/template` for templating, for more information on how to use them, refer to the official [documentation](https://golang.org/pkg/text/template).

When a journal has been initialized the `_layouts` directory will be empty, it will be up to you to create the necessary layout files. Below is an example of a layout file for a post.

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

Certain variables will be passed to each layout file during the publishing stage of a journal. Below is a table of the variables that will be passed to *all* rendered layouts.

| Variable | Purpose |
|----------|---------|
| `.Title` | The title of the journal that was set with `jrnl title`. |
| `.Categories` | A slice of all of the categories in the journal. |
| `.Pages` | A slice of all of the pages in the journal. |

Page, and post layouts will be passed the `.Page`, and `.Post` variables respectively. When each of these variables are passed to the layouts, their bodies will have already been rendered into HTML.

Index layouts will be passed the `.Posts` variable containing a slice of all of the posts to be rendered in that index. There are three main types of index layouts, an index time layout, a category index time layout, and a category index layout.

The index time layout will be passed a `.Time` variable containing the time of the current index. For example this would be used on an index that would be displaying all posts made on `2018-11-11`.

The category index time layout will be passed the `.Time` variable, and the `.Category` variable, and the category index will only be passed the `.Category` variable.

## Themes

Simply put, themes are just tarballs of the `_site/assets` directory, and of the `_layouts` directory. The current them in use is stored in the `_meta.yml` file. Themes can be created, saved, and deleted via the `jrnl theme` command.

You can check to see if a theme is in use by running `jrnl theme`.

```
blog $ jrnl theme
no theme being used
blog $
```

Creating, or saving a theme is done by running `jrnl theme save`. If you currently don't have a theme loaded then the `jrnl theme save` command will expect an argument to be passed. Otherwise the `jrnl theme save` command will overwrite the current theme in use.

Themes can be loaded by running `jrnl theme use`. This will take the tarball from the `_themes` directory, extrat it, and replace the necessary directories.

All available themes can be listed with `jrnl theme ls`, and themes can be removed with `jrnl theme rm`.

## Remotes

jrnl does not serve the content it generates, instead it simply copies everything in the `_site` directory to a remote. Remotes are managed via the `jrnl remote` command. To create a new remote simply run `jrnl remote set`.

```
blog $ jrnl remote set blog me@andrewpillar.com: -d
```

Here we're setting up a new remote called `blog`, that is pointing to the home directory on `andrewpillar.com`, and we're setting the `-d` flag on the command to tell jrnl that this is the default remote.

Now that this remote is setup the `_site` directory will be copied to that remove via SSH, since we specified an SCP URI, whenever we run `jrnl publish`.

When setting up a remote we can also specify which identity file to use for the SSH connection, along with the port to use, with the `-i`, and `-p` flags respectively.

```
blog $ jrnl remote set blog me@andrewpillar.com: -d -p 1234 -i ~/.ssh/id_blog
```

A remote can also be a location on disk, as well as a remote endpoint.

```
blog $ jrnl remote set local /var/www/blog
```

All remotes can be displayed with the `jrnl remote ls` command.

```
blog $ jrnl remote ls
blog  - me@andrewpillar.com: [default]
local - /var/www/blog
blog $
```

Remotes can be removed with `jrnl remote rm`.

```
blog $ jrnl remote rm local
```

## Publishing

A journal can be published via `jrnl publish`. This will take all pages, and posts that have been created, generate the HTML using the layouts specified, create any indexes, and copy them to the remote.

Different remotes can be specified via the `-r` flag on the `jrnl publish` command. If no remote is given via `-r` then the default remote will be used.

```
blog $ jrnl publish -r local
```

Drafts can also be published by setting the `-d` flag. This will only perform the conversion of Markdown to HTML, and write any indexes.

```
blog $ jrnl publish -d
```
