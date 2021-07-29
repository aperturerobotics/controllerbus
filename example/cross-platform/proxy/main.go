package main

import (
	"net/http"
	"os"
	"path"
)

func run() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	http.HandleFunc("/wasm_exec.js", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/javascript")
		http.ServeFile(rw, req, path.Join(wd, "../wasm_exec.js"))
	})
	http.HandleFunc("/example.wasm", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/wasm")
		http.ServeFile(rw, req, path.Join(wd, "../example.wasm"))
	})
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/":
		case "/index.html":
		case "/wasm_exec.html":
		default:
			return
		}
		http.ServeFile(rw, req, path.Join(wd, "../wasm_exec.html"))
	})

	// Start HTTP server
	os.Stderr.WriteString("listening on :5000\n")
	return http.ListenAndServe(":5000", nil)
}

func main() {
	if err := run(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
