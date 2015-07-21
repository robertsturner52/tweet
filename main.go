package main

import "net/http"

func init() {
	http.HandleFunc("/", handleIndex)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("assets/"))))

}
