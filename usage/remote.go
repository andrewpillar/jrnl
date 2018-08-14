package usage

var (
	Remote = `jrnl remote - Modify a remote

Usage:

  jrnl remote [command] [options...]

Commands:

  set  Set a new remote, or existing remote
  ls   List all remotes
  rm   Remove a remote

Options:

  --help  Display this usage message`

	RemoteSet = `jrnl remote set - Set a new, or existing remote

Usage:

  jrnl remote set [alias] [target]

Options:

  -d, --default  Set this remote as the default for publishing
  -p, --port     Change the port used to connecting to remote

  --help  Display this usage message`

	RemoteLs = `jrnl remote ls - List all remotes

Usage:

  jrnl remote ls

Options:

  --help  Display this usage message`

	RemoteRm = `jrnl remote rm - Remote a remote

Usage:

  jrnl remote rm [alias]

Options:

  --help  Display this usage message`
)
