package meta

var (
	index = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Title}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<div>
				<a href="{{$p.Href}}">{{$p.Title}}</a>
				<div>{{$p.CreatedAt.Format "Mon 2 Jan 2006"}}</div>
				<div>{{$p.Preview}}</div>
			</div>
		{{end}}
	</body>
</html>`

	dayIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>Posts from {{.Time.Format "Mon 2 Jan 2006"}} - {{.Title}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<div>
				<a href="{{$p.Href}}">{{$p.Title}}</a>
				<div>{{$p.CreatedAt.Format "Mon 2 Jan 2006"}}</div>
				<div>{{$p.Preview}}</div>
			</div>
		{{end}}
	</body>
</html>`

	monthIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>Posts from {{.Time.Format "Jan 2006"}} - {{.Title}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<div>
				<a href="{{$p.Href}}">{{$p.Title}}</a>
				<div>{{$p.CreatedAt.Format "Mon 2 Jan 2006"}}</div>
				<div>{{$p.Preview}}</div>
			</div>
		{{end}}
	</body>
</html>`

	yearIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>Posts from {{.Time.Format "2006"}} - {{.Title}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<div>
				<a href="{{$p.Href}}">{{$p.Title}}</a>
				<div>{{$p.CreatedAt.Format "Mon 2 Jan 2006"}}</div>
				<div>{{$p.Preview}}</div>
			</div>
		{{end}}
	</body>
</html>`

	categoryIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Category}} - {{.Title}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<div>
				<a href="{{$p.Href}}">{{$p.Title}}</a>
				<div>{{$p.CreatedAt.Format "Mon 2 Jan 2006"}}</div>
				<div>{{$p.Preview}}</div>
			</div>
		{{end}}
	</body>
</html>`

	categoryDayIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Category}} - {{.Title}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<div>
				<a href="{{$p.Href}}">{{$p.Title}}</a>
				<div>{{$p.CreatedAt.Format "Mon 2 Jan 2006"}}</div>
				<div>{{$p.Preview}}</div>
			</div>
		{{end}}
	</body>
</html>`

	categoryMonthIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Category}} - {{.Title}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<div>
				<a href="{{$p.Href}}">{{$p.Title}}</a>
				<div>{{$p.CreatedAt.Format "Mon 2 Jan 2006"}}</div>
				<div>{{$p.Preview}}</div>
			</div>
		{{end}}
	</body>
</html>`

	categoryYearIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Category}} - {{.Title}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<div>
				<a href="{{$p.Href}}">{{$p.Title}}</a>
				<div>{{$p.CreatedAt.Format "Mon 2 Jan 2006"}}</div>
				<div>{{$p.Preview}}</div>
			</div>
		{{end}}
	</body>
</html>`

	post = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Title}}</title>
	</head>
	<body>
		{{if .Post.HasCategory}}
			<h1>{{.Post.Category}} - {{.Post.Title}}</h1>
		{{else}}
			<h1>{{.Post.Title}}</h1>
		{{end}}
		<div>{{.Post.CreatedAt.Format "Mon 2 Jan 2006"}}</div>
		<div>{{.Post.Body}}</div>
	</body>
</html>`
)
