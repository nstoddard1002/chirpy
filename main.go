package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	//handle json api request
	chirpyServeMux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerChirpValidate)

	//create the HTTP server
	chirpyHTTPServer := &http.Server{
		Addr:    formatted_port_num,
		Handler: chirpyServeMux,
	}

	//print message to log
	log.Printf("Chirpy Server Listening on Port: %s", port_num)
	log.Fatal(chirpyHTTPServer.ListenAndServe())

}

func (cfg *apiConfig) handlerChirpValidate(w http.ResponseWriter, r *http.Request) {
	type chirp_msg struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirpychirp := chirp_msg{}
	err := decoder.Decode(&chirpychirp)
	if err != nil {
		log.Printf("error decoding JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	type error_response struct {
		Error_Msg string `json:"error"`
	}

	if len(chirpychirp.Body) > 140 {
		chirp_too_long_error := error_response{
			Error_Msg: "Chirp is too long",
		}

		datum, err1 := json.Marshal(chirp_too_long_error)
		if err1 != nil {
			log.Printf("error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(datum)
		return
	}

	split_body := strings.Split(chirpychirp.Body, " ")
	temp_word := ""
	for idx, word := range split_body {
		temp_word = strings.ToLower(word)
		if temp_word == "kerfuffle" || temp_word == "sharbert" || temp_word == "fornax" {
			split_body[idx] = "****"
		}
	}

	cleaned_chirp := strings.Join(split_body, " ")

	type success_response struct {
		Is_Valid     bool   `json:"valid"`
		Cleaned_Body string `json:"cleaned_body"`
	}

	chirp_valid_msg := success_response{
		Cleaned_Body: cleaned_chirp,
	}
	datum1, err2 := json.Marshal(chirp_valid_msg)
	if err2 != nil {
		log.Printf("error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(datum1)
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
