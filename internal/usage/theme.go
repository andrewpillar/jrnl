package usage

var (
	Theme = `jrnl theme - Manage the journal's themes

Usage:

  jrnl theme [command] [options...]

Commands:

  ls    List all journal themes
  save  Save the current theme
  use   Use a saved theme
  rm    Remove a theme

Options:

  --help Display this usage message

For more information on a command, and examples run 'jrnl theme [command] --help'`

	ThemeLs = `jrnl theme ls - List all journal themes

Usage:

  jrnl theme ls [options...]

Options:

  --help  Display this usage message

Examples:

  $ jrnl theme ls`

	ThemeSave = `jrnl theme save - Save the current theme

Usage:

  jrnl theme save [name] [options...]

Options:

  --help  Display this usage message

Examples:

  Save the current theme as a new theme:

    $ jrnl theme save tty

  Overwrite the current theme:

    $ jrnl theme save`

	ThemeUse = `jrnl theme use - Use a saved theme

Usage:

  jrnl theme use [name] [options...]

Options:

  --help  Display this usage message

Examples:

  $ jrnl theme use tty`

	ThemeRm = `jrnl theme rm - Remove a theme

Usage:

  jrnl theme rm [name...] [options...]

Options:

  --help  Display this usage message

Examples:

  Remove a single theme:

    $ jrnl theme rm tty

  Remove multiple themes:

    $ jrnl theme rm tty midnight`
)
