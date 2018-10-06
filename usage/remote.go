package usage

var (
	Remote = `jrnl remote - Manage the journal's remotes

Usage:

  jrnl remote [command] [options...]

Commands:

  ls   List all journal remotes
  set  Set a remote
  rm   Remove a remote

Options:

  --help  Display this usage message

For more information on a command run 'jrnl remote [command] --help'`

	RemoteLs = `jrnl remote ls - List all journal remotes

Usage:

  jrnl remote ls [options...]

Options:

  -v, --verbose  Display more information about each remote

  --help  Display this usage message`

	RemoteSet = `jrnl remote set - Set a remote

Usage:

  jrnl remote set [remote] [target] [options...]

Options:

  -d, --default          Set the remote as a default
  -i, --identity=<file>  The SSH identity file for the remote
  -p, --port=<port>      The SSH port of the remote

  --help Display this usage message`

	RemoteRm = `jrnl remote rm - Remove a remot

Usage:

  jrnl remote rm [remotes...]

Options:

  --help  Display this usage message`
)
