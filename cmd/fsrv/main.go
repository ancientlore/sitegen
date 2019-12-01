package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	var (
		fAddr = flag.String("addr", ":8080", "Address to serve on.")
		fDir  = flag.String("dir", ".", "Folder to serve.")
	)
	flag.Parse()
	err := http.ListenAndServe(*fAddr, http.FileServer(http.Dir(*fDir)))
	if err != nil {
		log.Print(err)
	}
}
