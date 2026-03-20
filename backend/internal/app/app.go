package app

import (
	"log"
	"net/http"

	"swift-gopher/internal/dbconn"
)

func Run() {

	// test
	r := http.NewServeMux()

	r.HandleFunc("GET /hello/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.PathValue("id")

		var name string
		err := dbconn.GetDB().QueryRow("select name from test where id = $1", id).Scan(&name)
		if err != nil {
			w.Write([]byte(`{"error":"` + err.Error() + `"}`))
			return
		}

		w.Write([]byte(`{"hello":"` + name + `"}`))
	})

	log.Fatal(http.ListenAndServe(":8080", r))
	// test

}
