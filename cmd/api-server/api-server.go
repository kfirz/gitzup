package main

import (
	"html/template"
	"log"
	"net/http"
)

var tmpl = template.Must(template.New("index").Parse(`
<html>
	<head>
		<title>API Server</title>
	</head>
	<body>
		<h1>API Server</h1>
		<p>This is the API Server.</p>
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
