package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"goji.io"
	"goji.io/pat"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//
func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{message: %q}", message)
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}

// person模型
type Person struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func main() {
	session, err := mgo.Dial(os.Getenv("MONGO_URL"))
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	// handler
	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/people"), allPeople(session))
	mux.HandleFunc(pat.Post("/people"), addPerson(session))
	mux.HandleFunc(pat.Get("/people/:name"), personByName(session))
	mux.HandleFunc(pat.Put("/people/:name"), updatePerson(session))
	//mux.HandleFunc(pat.Delete("/people/:name"), deletePerson(session))
	http.ListenAndServe("localhost:8888", mux)
}

func allPeople(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		c := session.DB("go-test").C("people")

		var person []Person
		err := c.Find(bson.M{}).All(&person)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all persons: ", err)
			return
		}

		respBody, err := json.MarshalIndent(person, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}

func addPerson(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var person Person
		err := json.NewDecoder(r.Body).Decode(&person)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		err = session.DB("go-test").C("people").Insert(person)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed insert person: ", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", r.URL.Path+"/"+person.Name)
		w.WriteHeader(http.StatusCreated)
	}
}

func personByName(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()
		name := pat.Param(r, "name")

		var person Person
		err := session.DB("go-test").C("people").Find(bson.M{"name": name}).One(&person)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed find person: ", err)
			return
		}

		if person.Name == "" {
			ErrorWithJSON(w, "person not found", http.StatusNotFound)
			return
		}

		respBody, err := json.MarshalIndent(person, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}

func updatePerson(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		name := pat.Param(r, "name")

		var person Person
		err := json.NewDecoder(r.Body).Decode(&person)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		err = session.DB("go-test").C("people").Update(bson.M{"name": name}, &person)
		if err != nil {
			switch err {
			default:
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed update person: ", err)
				return
			case mgo.ErrNotFound:
				ErrorWithJSON(w, "person not found", http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func deletePerson(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
