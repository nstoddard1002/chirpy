package main

import (
	"log"
	"net/http"
)

func main() {
	//create the "ServeMux"
	chirpyServeMux := http.NewServeMux()

	//set the port number
	port_num := "8080"

	//address string formatting
	formatted_port_num := ":" + port_num

	//create the HTTP server
	chirpyHTTPServer := &http.Server{
		Addr:    formatted_port_num,
		Handler: chirpyServeMux,
	}

	//print message to log
	log.Printf("Chirpy Server Listening on Port: %s", port_num)
	log.Fatal(chirpyHTTPServer.ListenAndServe())

}
