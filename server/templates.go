package main

import (
	"html/template"
)

const inUse = `<html>
    <head>
        <link href="/static/index.css" rel="stylesheet">
    </head>
    <body>
        <main>
            <h1>This theater is currently in use.</h1>
	    <h2><a href="/projectionist/{{ . }}">{{ . }}</a> should be free!</h2>
        </main>
    </body>
</html>`

var inUsePage = template.Must(template.New("InUse").Parse(inUse))
