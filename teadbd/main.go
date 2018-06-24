package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.com/hokiegeek/teadb"
)

func main() {
	portPtr := flag.Int("port", 8888, "Specify the port to use")
	flag.Parse()
	fmt.Printf("Serving on port: %d\n", *portPtr)

	r := mux.NewRouter()
	r.HandleFunc("/tea/{id:[0-9]+}", teaHandler).Methods("GET", "POST", "PUT")
	r.HandleFunc("/teas", getAllTeasHandler).Methods("GET")
	r.HandleFunc("/entry", entryHandler).Methods("GET", "POST", "PUT")

	http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), r)
}

func teaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch r.Method {
	case http.MethodGet:
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		tea, err := teadb.GetTeaByID(id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		postJSON(w, http.StatusOK, tea)
	case http.MethodPost:
		// TODO Create new TEA
	case http.MethodPut:
		// TODO Update TEA
	}
}

func getAllTeasHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	teas, err := teadb.GetAllTeas()
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error)
	}
	postJSON(w, http.StatusOK, teas)
}

func entryHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ENTRY: %v\n", "TODO")
}

func postJSON(w http.ResponseWriter, httpStatus int, send interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(send); err != nil {
		panic(err)
	}
}
