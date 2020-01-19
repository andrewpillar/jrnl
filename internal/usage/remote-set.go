package usage

var RemoteSet = `jrnl remote-set - Set the journal's remote

Usage:

  jrnl remote-set [remote] [options...]

Options:

  --help  Display this usage message

Examples:

  Set the journal's remote to a local filesystem path:

    $ jrnl remote-set /var/www/my-site

  Set the journal's remote to a remote SCP path:

    $ jrnl remote-set me@my-blog.com:/var/www`
