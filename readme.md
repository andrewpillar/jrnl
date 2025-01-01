# jrnl

jrnl is a simple static site generator. It takes posts and pages written in
Markdown, transforms them to HTML, and copies those HTML files over to a
remote.

* [Quick start](#quick-start)
* [Directory structure](#directory-structure)
* [Pages and posts](#pages-and-posts)
* [Front matter](#front-matter)
* [Dynamic data](#dynamic-data)
* [Layouts](#layouts)
* [Themes](#themes)
* [Remote](#remote)
* [Publishing](#publishing)
* [Atom and RSS feeds](#atom-and-rss-feeds)

## Quick start

Clone the repository and install jrnl via `make install`. This will install
it to `$GOPATH/bin`,

    $ git clone https://github.com/andrewpillar/jrnl
    $ cd jrnl && make install

Once installed a new jrnl can be created with `jrnl init`,

    $ jrnl init my-blog

this will initialize everything needed for a jrnl within the `my-blog`
directory. If no argument is given, then a new jrnl will be initialized in the
current directory.

Once the jrnl has been initialized the directory contents will look like this,

    _assets/
    _layouts/
    _pages/
    _posts/
    _site/
    jrnl.yml

The site name and author information can be set in the `jrnl.yml` file,

    author:
        name: Joe Bloggs
        email: joe.bloggs@localhost

    site:
      title: My blog
      atom: atom.xml
      rss: rss.xml

A post can be created via `jrnl post`,

    $ jrnl post "Introducing jrnl"

jrnl will open up the editor specified via `EDITOR` for writing the post. The
top of the post will be pre-populated with a block of YAML which will contain
metadata about the post being written,

    ---
    title: Introducing jrnl
    layout: page
    createdAt: "2006-01-02 15:04:05"
    updatedAt: "2006-01-02 15:04:05"
    ---
    jrnl is a simple static site generator.

The layout specifies the template to use when generating the post. Layouts can
be found in the `_layouts/` directory. The default layout of `page` is used when
creating a post. The layout can be specified via the `-l` flag to the `post`
command,

    $ jrnl post -l post "Introducing jrnl"

Layouts utilize Go's [text/template][0] library for templating.

[0]: https://pkg.go.dev/text/template

Before a jrnl can be published a remote needs to be set in the `jrnl.yml` file.
This needs to be a URL. A remote can either be a location on disk, or an SSH
connection. For this example, a location on disk will be used,

    remote: file://remote

With the remote set, the jrnl can be published vis `jrnl publish`,

    $ jrnl publish

This will build the site and place the files in the `_site/` directory and copy
them to the remote.

## Directory structure

The following directories are used by jrnl,

* `_assets/` - Contains assets that are part of the theme. Directory contents is
  also copied to the `_sites/` directory during publishing.
* `_layouts/` - Contains the templates used for building posts and pages.
* `_pages/` - Contains the Markdown files for the pages.
* `_posts/` - Contains the Markdown files for the posts.
* `_site/` - Where the static files for the site are put during publishing.
* `jrnl.yml` - The configuration file for the jrnl.

## Pages and posts

Under the hood, pages and posts are the same. To jrnl, the only difference is
that posts are considered "dynamic" content, whereas pages are not. Posts make
use of the `createdAt` and `updatedAt` fields in the front-matter.

## Front matter

Front matter is the block of YAML that sits at the top of pages and posts. This
contains the metadata about the page or post.

## Dynamic data

Dynamic data can be populated into pages and posts. This is done by specifying
a YAML file from which to pull data from in the front matter of the page or
post,

    ---
    title: CV
    data: cv.yml
    layouts: page
    ---

assume the contents of `cv.yml` looks like this,

    Jobs:
    - Name: SRE
      StartDate: 2024
      Company: ACME
      Duties:
      - Manage cloud infrastructure via Terraform
      - Build reliable CI/CD pipelines for deploying infrastructure changes
      - Architect multi-region failover

this can then be populated inside a page like so,

    ---
    title: CV
    data: cv.yml
    layout: page
    ---
    {{range .Jobs}}
    ## {{.Role}} ({{.StartDate}} - {{if .EndDate}}{{.EndDate}}{{else}}Present{{end}})

    **{{.Company}}**

    {{range .Duties}}
    * {{.}}
    {{end}}
    {{end}}

this allows for populating structured and repetive data with ease in a page or
a post.

## Layouts

Layouts are template files that define how a page or post will look once
published. jrnl uses Go's [text/template][0] library for templating. When jrnl
is initialized the `_layouts/` directory will be populated with the default
`home` and `page` template files.

jrnl makes use of the `home` layout to determine what the home page of the
site will end up looking like. If this is removed, then any attempts to publish
the jrnl will fail.

During templating, the following data is passed to each page,

    type PageData struct {
        Site     Site
        Author   Author
        Title    string
        Content  string
        Posts    []SitePage
        Blogroll []SitePage
    }

## Themes

Themes in jrnl are tarballs that contain the contents of the `_assets/` and
`_layouts/` directories. The current theme being used is set in the `jrnl.yml`
file. All themes are stored in the user's configuration dir.

## Remote

Each jrnl has a remote. A remote is where the contents of the `_site/` directory
is copied to. This can either be a location on disk, or an SSH URL. Below is
what the SSH URL should look like,

    ssh://<username>@<host>:[port]/<path>[?identity=<path>]

the `identity` parameter should be the path to the private SSH key to use for
the handshake.

## Publishing

A jrnl is published via the `jrnl publish` command. This will template the pages
and posts, and copy the contents of the `_assets/` directory into the `_site/`
directory.

The `-d` flag will prevent the built website from being copied to the remote.

The `-v` flag will print out the files that have been built, and which files
will be copied to the remote.

Each site path that is created will be stored in the jrnl's cache. This is
checked each time `publish` is run to ensure that only the files that have
changed will be copied to the remote.

## Atom and RSS feeds

Atom and RSS feeds are generated each time a jrnl is published. The files that
these are written to are specified via the `atom` and `rss` fields in the
`jrnl.yml` file. If these are not specified, then no feeds are generated.
