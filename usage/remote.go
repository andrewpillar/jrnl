package usage

var Remote = `jrnl remote - Modify a remote

Usage:

  jrnl remote [command] [options...]

Commands:

  ls   List available remotes
  set  Set a remote
  rm   Remove a remote

Options:

  --help  Display this usage message

For more information on a command run 'jrnl remote [command] --help`

var RemoteLs = `jrnl remote ls - List available remotes

Usage:

  jrnl remote ls [options...]

Options:

  --help  Display this usage message`

var RemoteSet = `jrnl remote set - Set a remote

Usage:

  jrnl remote set [alias] [target] [options...]

Options:

  -d, --default          Set the remote as the default
  -i, --identity=<file>  The SSH identity file for the remote
  -p, --port=<port>      The port for the remote

  --help  Display this usage message`

var RemoteRm = `jrnl remote rm - Remove a remot

Usage:

  jrnl remote rm [aliases...]

Options:

  --help  Display this usage message`
