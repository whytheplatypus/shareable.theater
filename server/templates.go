package main

import (
	"html/template"
)

const inUse = `<html>
    <head>
        <style>
            main {
                height: 100%;
                display: flex;
                justify-content: center;
                align-items: center;
                flex-direction: column;
            }
        </style>
    </head>
    <body>
        <main>
            <h1>This theater is currently in use.</h1>
	    <h2><a href="/booth/{{ . }}">{{ . }}</a> should be free!</h2>
        </main>
    </body>
</html>`

var inUsePage = template.Must(template.New("InUse").Parse(inUse))
