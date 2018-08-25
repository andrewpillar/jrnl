package meta

var (
	index = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Title}}</title>
	</head>
	<body>
	</body>
</html>`

	dayIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Title}}</title>
	</head>
	<body>
	</body>
</html>`

	monthIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Title}}</title>
	</head>
	<body>
	</body>
</html>`

	yearIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Title}}</title>
	</head>
	<body>
	</body>
</html>`

	categoryIndex = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Title}}</title>
	</head>
	<body>
	</body>
</html>`

	post = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title>{{.Title}}</title>
	</head>
	<body>
		<h1>{{.Post.Title}}</h1>
		<div>{{.Post.Body}}</div>
	</body>
</html>`
)
