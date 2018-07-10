package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gitlab.com/hokiegeek.net/teadb"
)

func main() {
	portPtr := flag.Int("port", 80, "Specify the port to use")
	flag.Parse()
	fmt.Printf("Serving on port: %d\n", *portPtr)

	r := mux.NewRouter()
	r.HandleFunc("/teas", getAllTeasHandler).Methods("HEAD", "GET", "OPTIONS")
	r.HandleFunc("/tea/{id:[0-9]+}", teaHandler).Methods("HEAD", "GET", "POST", "PUT", "DELETE", "OPTIONS")
	r.HandleFunc("/tea/{teaid:[0-9]+}/entry", entryHandler).Methods("HEAD", "GET", "POST", "PUT", "OPTIONS")

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), handlers.CORS(originsOk, headersOk, methodsOk)(r))
}

func cors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "HEAD, POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With")
}

func getAllTeasHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s /teas [%s]\n", r.Method, r.RemoteAddr)

	cors(&w)

	if r.Method == http.MethodOptions {
		return
	}

	teas, err := teadb.GetAllTeas()
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		postJSON(w, r, teas)
	}
}

func teaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	log.Printf("%s /tea/%s [%s]\n", r.Method, vars["id"], r.RemoteAddr)

	cors(&w)

	if r.Method == http.MethodOptions {
		return
	}

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		switch r.Method {
		case http.MethodHead:
		case http.MethodGet:
			tea, err := teadb.GetTeaByID(id)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			} else {
				postJSON(w, r, tea)
			}
		case http.MethodPost:
			// Create new TEA
			_, err := teadb.GetTeaByID(id)
			if err == nil {
				w.WriteHeader(http.StatusConflict)
			} else {
				if tea, err := readTea(w, r); err == nil {
					if err = teadb.CreateTea(tea); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.WriteHeader(http.StatusCreated)
					}
				}
			}
		case http.MethodPut:
			// Update existing TEA
			_, err := teadb.GetTeaByID(id)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			} else {
				if tea, err := readTea(w, r); err == nil {
					if err = teadb.UpdateTea(tea); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.WriteHeader(http.StatusOK)
					}
				}
			}
		case http.MethodDelete:
			// Delete existing TEA
			if _, err := teadb.GetTeaByID(id); err != nil {
				w.WriteHeader(http.StatusNotFound)
			} else {
				if err = teadb.DeleteTea(id); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	}
}

func entryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	log.Printf("%s /tea/%s/entry [%s]\n", r.Method, vars["teaid"], r.RemoteAddr)

	cors(&w)

	if r.Method == http.MethodOptions {
		return
	}

	teaid, err := strconv.Atoi(vars["teaid"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tea, err := teadb.GetTeaByID(teaid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodHead:
	case http.MethodGet:
		postJSON(w, r, tea)
	case http.MethodPost:
		if entry, err := readEntry(w, r); err == nil {
			if err = teadb.CreateEntry(tea.ID, entry); err != nil {
				http.Error(w, "error creating new entry", http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusCreated)
			}
		}
	case http.MethodPut:
		// Update existing tea entry
		if entry, err := readEntry(w, r); err == nil {
			if err = teadb.UpdateEntry(tea.ID, entry); err != nil {
				http.Error(w, "error updating entry", http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}
	}
}

func postJSON(w http.ResponseWriter, r *http.Request, payload interface{}) {
	if JSON, err := json.Marshal(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		requestedSum := r.Header.Get("If-None-Match")
		sum := checksum(JSON)

		if requestedSum == sum {
			w.WriteHeader(http.StatusNotModified)
		} else {
			w.Header().Set("ETag", sum)
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")

				if err = json.NewEncoder(w).Encode(payload); err != nil {
					log.Printf("Error sending payload: %v", err)
					http.Error(w, "error sending payload", http.StatusInternalServerError)
				}
			} else {
				w.Header().Add("Content-Length", "0")
				w.WriteHeader(http.StatusOK)
			}
		}
	}
}

func checksum(body []byte) string {
	hash := md5.Sum(body)
	return hex.EncodeToString(hash[:])
}

func readEntry(w http.ResponseWriter, r *http.Request) (entry teadb.TeaEntry, err error) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	// TODO: need some way to validate this

	if err = json.Unmarshal(body, &entry); err != nil {
		log.Printf("Could not unmarshal: %s\n", err)
		http.Error(w, "can't read entry", http.StatusUnprocessableEntity)
	}

	return
}

func readTea(w http.ResponseWriter, r *http.Request) (tea teadb.Tea, err error) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	// TODO: need some way to validate this

	if err = json.Unmarshal(body, &tea); err != nil {
		log.Printf("Could not unmarshal: %s\n", err)
		http.Error(w, "can't read tea", http.StatusUnprocessableEntity)
	}

	return
}
