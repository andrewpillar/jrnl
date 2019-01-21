package usage

var Post = `jrnl post - Create a post

Usage:

  jrnl post [title] [options...]

Options:

  -c, --category  The category for the new post

  --help  Display this usage message

Examples:

  Create a standalone post:

    $ jrnl post "Introducing jrnl"

  Create a post in a category:

    $ jrnl post -c "Programming" "Go, 101"

  Create a post in a sub-category:

    $ jrnl post -c "TV Shows / The Leftovers" "Penguin One, Us Zero"`
