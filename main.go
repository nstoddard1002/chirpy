package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	//create the "ServeMux"
	chirpyServeMux := http.NewServeMux()

	//set the port number
	port_num := "8080"

	//filepath variable
	filepath := "."

	//address string formatting
	formatted_port_num := ":" + port_num

	//create the apiConfig struct object
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}

	//handle /app, now with middlewaremetricsinc function
	chirpyServeMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepath)))))
	//handle /healthz
	chirpyServeMux.HandleFunc("GET /api/healthz", handlerReadiness)
	//handle the metrics request
	chirpyServeMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	//handle the reset request
	chirpyServeMux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	//handle bird image
	chirpyServeMux.Handle("/assets/logo.png", http.FileServer(http.Dir(filepath)))

	//create the HTTP server
	chirpyHTTPServer := &http.Server{
		Addr:    formatted_port_num,
		Handler: chirpyServeMux,
	}

	//print message to log
	log.Printf("Chirpy Server Listening on Port: %s", port_num)
	log.Fatal(chirpyHTTPServer.ListenAndServe())

}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
	<html>
	
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>
	`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits reset to 0")))
}
