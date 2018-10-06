package usage

var (
	Theme = `jrnl theme - Manage the journal's themes

Usage:

  jrnl theme [command] [options...]

Commands:

  ls    List available themes
  save  Save the current theme
  use   Use a theme
  rm    Remove a theme

Options:

  --help  Display this usage message

For more information on a command run 'jrnl theme [command] --help'`

	ThemeLs = `jrnl theme ls - List available themes

Usage:

  jrnl theme ls [options...]

Options:

  --help  Display this usage message`

	ThemeSave = `jrnl theme save - Save the current theme

Usage:

  jrnl theme save [theme] [options...]

Options:

  --help  Display this usage message`

	ThemeUse = `jrnl theme use - Use a theme

Usage:

  jrnl theme use [theme] [options...]

Options:

  --help  Display this usage message`

	ThemeRm = `jrnl theme rm - Remove a theme

Usage:

  jrnl theme rm [themes...] [options...]

Options:

  --help  Display this usage message`
)
