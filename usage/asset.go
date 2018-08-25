package usage

var Asset = `jrnl asset - Modify the journal's assets

Usage:

  jrnl asset [command] [options...]

Commands:

  ls    List available assets
  add   Add a new asset
  edit  Edit an asset
  rm    Remove an asset

Options:

  --help  Display this usage message

For more information on a command run 'jrnl asset [command] --help`

var AssetLs = `jrnl asset ls - List available assets

Usage:

  jrnl asset ls [options...]

Options:

  --help  Display this usage message`

var AssetAdd = `jrnl asset add - Add a new asset

Usage:

  jrnl asset add [asset] [options...]

Options:

  -f, --file=<file>  The file on disk to add as an asset
  -d, --dir=<dir>    The directory under the assets directory to put the asset

  --help  Display this usage message`

var AssetEdit = `jrnl asset edit - Edit an asset

Usage:

  jrnl asset edit [asset]

Options:

  --help  Display this usage message`

var AssetRm = `jrnl asset rm - Remove an asset

Usage:

  jrnl asset rm [assets...]

Options:

  --help  Display this usage message`
