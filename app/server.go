package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var (
	httpport = flag.String("httpport", ":8080", "httpport")
	version  = "v0.0.17"
)

const ()

func fronthandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/ called")
	fmt.Fprint(w, "ok")
}

func main() {

	flag.Parse()

	fmt.Printf("VERSION: %s\n", version)

	if *httpport == "" {
		fmt.Fprintln(os.Stderr, "missing -httpport flag (:8080)")
		flag.Usage()
		os.Exit(2)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", fronthandler)

	fmt.Println("starting server")
	err := http.ListenAndServe(*httpport, r)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
