package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world,  alpoGame starts -> %s", r.URL.Path)
}

func main() {
	fmt.Println("run server - alpoGame welcome")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8088", nil))
}
