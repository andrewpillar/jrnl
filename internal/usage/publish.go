package usage

var Publish = `jrnl publish - Publish the journal

Usage:

  jrnl publish [options...]

Options:

  -a, --atom   The file to write the Atom feed to
  -d, --draft  Don't copy the published site to the remote
  -r, --rss    The file to write the RSS feed to

  --help  Display this usage message

Examples:

  Publish a draft:

    $ jrnl publish -d

  Publish with an Atom feed:

    $ jrnl publish -a _site/atom.xml

  Publish with an RSS feed:

    $ jrnl publish -r _site/rss.xml`
