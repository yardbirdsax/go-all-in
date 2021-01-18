package main

import (
	"fmt"
	"log"
	"net/http"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request is: %v",r.Body)
	fmt.Fprintf(w, "Hello World!")
}

func handleRequests() {
	http.HandleFunc("/",helloWorld)
	log.Fatal(http.ListenAndServe(":10000",nil))
}

func main() {
	handleRequests()
}