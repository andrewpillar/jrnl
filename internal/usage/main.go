package usage

var (
	Commands = map[string]string{
		"init":       Init,
		"title":      Title,
		"page":       Page,
		"post":       Post,
		"ls":         Ls,
		"edit":       Edit,
		"rm":         Rm,
		"remote-set": RemoteSet,
		"publish":    Publish,
		"theme":      Theme,
		"theme-ls":   ThemeLs,
		"theme-save": ThemeSave,
		"theme-use":  ThemeUse,
		"theme-rm":   ThemeRm,
		"gen-style":  GenStyle,
	}

	Jrnl = `jrnl - A simple static site generator

Usage:

  jrnl [command] [options...]

Commands:

  init        Initialize a new journal
  title       Set the journal's title
  page        Create a page
  post        Create a post
  ls          List all journal pages, and posts
  edit        Edit a page, or post
  rm          Remove a page, or post
  remote-set  Set the journal's remote
  publish     Publish the journal
  theme       Manage the journal's themes
  gen-style   Generate CSS for the journal's syntax highlighting

Options:

  --help  Display this usage message

For more information on a command, and examples run 'jrnl [command] --help'`
)
