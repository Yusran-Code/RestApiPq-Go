package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var db *sql.DB

func main() {
	var err error
	connStr := "postgres://postgres:%40yusran@db.jtsuknwzkkmamvncosny.supabase.co:5432/postgres?sslmode=require"

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("gagal",err)
	}
	fmt.Println("Succes DB Connection")

	router := mux.NewRouter()
	router.HandleFunc("/example", getItems).Methods("GET")
	router.HandleFunc("/example/{id}", getItem).Methods("GET")
	router.HandleFunc("/example", createItem).Methods("POST")
	router.HandleFunc("/example/{id}", updateItem).Methods("PUT")
	router.HandleFunc("/example/{id}", deleteItem).Methods("DELETE")

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// GET /items
func getItems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, description FROM example")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var it Item
		err := rows.Scan(&it.ID, &it.Name, &it.Description)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		items = append(items, it)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// GET /items/{id}
func getItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}

	var it Item
	err = db.QueryRow("SELECT id, name, description FROM example WHERE id = $1", idInt).
		Scan(&it.ID, &it.Name, &it.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Item not found", 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(it)
}

// POST /items
func createItem(w http.ResponseWriter, r *http.Request) {
	var it Item
	err := json.NewDecoder(r.Body).Decode(&it)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	

	err = db.QueryRow(
		"INSERT INTO example (name, description) VALUES ($1, $2) RETURNING id",
		it.Name, it.Description,
	).Scan(&it.ID)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(it)
}

// PUT /items/{id}
func updateItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}

	var it Item
	err = json.NewDecoder(r.Body).Decode(&it)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}



	res, err := db.Exec("UPDATE example SET name=$1, description=$2 WHERE id=$3", it.Name, it.Description, idInt)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	count, err := res.RowsAffected()
	if err != nil || count == 0 {
		http.Error(w, "Item not found or no change", 404)
		return
	}

	it.ID = idInt
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(it)
}

// DELETE /items/{id}
func deleteItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}

	res, err := db.Exec("DELETE FROM example WHERE id=$1", idInt)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	count, err := res.RowsAffected()
	if err != nil || count == 0 {
		http.Error(w, "Item not found", 404)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
