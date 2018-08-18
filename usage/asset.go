package usage

var Asset = `jrnl asset - Modify the site's assets

Usage:

  jrnl asset [command] [options...]

Commands:

  ls    List all assets
  add   Add a new asset
  edit  Edit an asset
  rm    Remove an asset

Options:

  --help  Display this usage message`

var AssetLs = `jrnl asset ls - List all assets

Usage:

  jrnl asset ls

Options:

  --help  Display this usage message`

var AssetAdd = `jrnl asset add - Add a new asset

Usage:

  jrnl asset add [asset] [options...]

Options:

  -f, --file=<file>   The target file on disk to add as an asset
  -t, --target=<dir>  The target directory to place the asset

  --help  Display this usage message`

var AssetEdit = `jrnl asset edit - Edit an asset

Usage:

  jrnl asset edit [asset]

Options:

  --help  Dislay this usage message`

var AssetRm = `jrnl asset rm - Remove an asset

Usage:

  jrnl asset rm [assets...]

Options:

  --help  Display this usage message`
