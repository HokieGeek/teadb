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

	db, err := teadb.New()
	if err != nil {
		panic(err)
	}

	cache, err := NewCache()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/teas",
		func(w http.ResponseWriter, r *http.Request) {
			getAllTeasHandler(w, r, db, cache)
		}).Methods("HEAD", "GET", "OPTIONS")

	r.HandleFunc("/tea/{id:[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			teaHandler(w, r, db, cache)
		}).Methods("HEAD", "GET", "POST", "PUT", "DELETE", "OPTIONS")

	r.HandleFunc("/tea/{teaid:[0-9]+}/entry",
		func(w http.ResponseWriter, r *http.Request) {
			entryHandler(w, r, db, cache)
		}).Methods("HEAD", "POST", "PUT", "OPTIONS")

	r.HandleFunc("/tea/{teaid:[0-9]+}/entry/{entryid:[0-9]*}",
		func(w http.ResponseWriter, r *http.Request) {
			entryHandler(w, r, db, cache)
		}).Methods("HEAD", "GET", "PUT", "DELETE", "OPTIONS")

	originsOk := handlers.AllowedOrigins([]string{"*"})
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "If-None-Match"})
	exposedOk := handlers.ExposedHeaders([]string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Etag"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), handlers.CORS(originsOk, headersOk, methodsOk, exposedOk)(r))
}

func getAllTeasHandler(w http.ResponseWriter, r *http.Request, db *teadb.GcpClient, cache *Cache) {
	log.Printf("%s /teas [%s]\n", r.Method, r.RemoteAddr)

	if r.Method == http.MethodOptions {
		return
	}

	var teas []teadb.Tea
	var err error
	if cache.AllTeasValid() {
		teas = cache.GetAllTeas()
		log.Println("Using cache")
	} else {
		teas, err = db.GetAllTeas()
		if err == nil {
			cache.CacheAllTeas(teas)
			log.Println("Retrieved and cached")
		}
	}

	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		postJSON(w, r, teas)
	}
}

func teaHandler(w http.ResponseWriter, r *http.Request, db *teadb.GcpClient, cache *Cache) {
	vars := mux.Vars(r)

	log.Printf("%s /tea/%s [%s]\n", r.Method, vars["id"], r.RemoteAddr)

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
			var tea teadb.Tea
			var err error
			if cache.IsTeaValid(id) {
				tea = cache.GetTea(id)
			} else {
				tea, err = db.GetTeaByID(id)
				if err == nil {
					cache.SetTea(tea)
				}
			}

			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			} else {
				postJSON(w, r, tea)
			}
		case http.MethodPost:
			// Create new TEA
			_, err := db.GetTeaByID(id)
			if err == nil {
				w.WriteHeader(http.StatusConflict)
			} else {
				if tea, err := readTea(w, r); err == nil {
					if err = db.CreateTea(tea); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						cache.Invalidate()
						w.WriteHeader(http.StatusCreated)
					}
				}
			}
		case http.MethodPut:
			// Update existing TEA
			_, err := db.GetTeaByID(id)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			} else {
				if tea, err := readTea(w, r); err == nil {
					if err = db.UpdateTea(tea); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						cache.Invalidate()
						w.WriteHeader(http.StatusOK)
					}
				}
			}
		case http.MethodDelete:
			// Delete existing TEA
			if _, err := db.GetTeaByID(id); err != nil {
				w.WriteHeader(http.StatusNotFound)
			} else {
				if err = db.DeleteTea(id); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					cache.Invalidate()
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	}
}

func entryHandler(w http.ResponseWriter, r *http.Request, db *teadb.GcpClient, cache *Cache) {
	vars := mux.Vars(r)

	log.Printf("%s /tea/%s/entry/%s [%s]\n", r.Method, vars["teaid"], vars["entryid"], r.RemoteAddr)

	if r.Method == http.MethodOptions {
		return
	}

	teaid, err := strconv.Atoi(vars["teaid"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tea, err := db.GetTeaByID(teaid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var entryid int64
	if len(vars["entryid"]) > 0 && r.Method != http.MethodPost && r.Method != http.MethodPut {
		entryid, err = strconv.ParseInt(vars["entryid"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	switch r.Method {
	case http.MethodHead:
	case http.MethodGet:
		var entry teadb.TeaEntry
		for _, teaEntry := range tea.Entries {
			if entryid == teaEntry.Datetime.UnixNano() {
				entry = teaEntry
				break
			}
		}
		postJSON(w, r, entry)
	case http.MethodPost:
		if entry, err := readEntry(w, r); err == nil {
			if err = db.CreateEntry(tea.ID, entry); err != nil {
				http.Error(w, "error creating new entry", http.StatusInternalServerError)
			} else {
				cache.Invalidate()
				w.WriteHeader(http.StatusCreated)
			}
		}
	case http.MethodPut:
		// Update existing tea entry
		if entry, err := readEntry(w, r); err == nil {
			if err = db.UpdateEntry(tea.ID, entry); err != nil {
				http.Error(w, "error updating entry", http.StatusInternalServerError)
			} else {
				cache.Invalidate()
				w.WriteHeader(http.StatusOK)
			}
		}
	case http.MethodDelete:
		for i, teaEntry := range tea.Entries {
			log.Printf("%d == %d\n", entryid, teaEntry.Datetime.Unix())
			if entryid == teaEntry.Datetime.Unix() {
				log.Printf("   found it! %d: %v\n", i, teaEntry)
				tea.Entries = append(tea.Entries[:i], tea.Entries[i+1:]...)
				break
			}
		}

		if err = db.UpdateTea(tea); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			cache.Invalidate()
			w.WriteHeader(http.StatusOK)
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
			w.Header().Set("Etag", sum)
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
