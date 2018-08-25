package main

import (
	"html/template"
	"log"
	"net/http"
)

var tmpl = template.Must(template.New("index").Parse(`
<html>
	<head>
		<title>Webhooks Server</title>
	</head>
	<body>
		<h1>Webhooks Server</h1>
		<p>This is the web-hooks server.</p>
	</body>
</html>`))

func handleRequest(w http.ResponseWriter, req *http.Request) {
	tmpl.Execute(w, req.FormValue("s"))
}

func main() {
	http.Handle("/", http.HandlerFunc(handleRequest))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
