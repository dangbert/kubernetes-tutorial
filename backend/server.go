// toy server implementing demo endpoints

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// routing
	http.HandleFunc("/fruit", getFruit)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<b>Hello world</b>`)
		fmt.Printf("saying hello\n")
	})

	fmt.Printf("running on port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}

type Fruit struct {
	ID   int    `json:"ID"`
	Name string `json:"name"`
}

// craft an http response containing a json list
func getFruit(w http.ResponseWriter, r *http.Request) {
	items := []Fruit{{ID: 0, Name: "Apple"}, {ID: 1, Name: "Banana"}}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
	fmt.Printf("returning %d fruits\n", len(items))
}
