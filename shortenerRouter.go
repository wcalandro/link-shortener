package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/wcalandro/base62"
	redistore "gopkg.in/boj/redistore.v1"
)

func shortenerRouter(store *redistore.RediStore) chi.Router {
	// MySQL database
	db := DB

	r := chi.NewRouter()

	// Redirect to site on root
	websiteURL := os.Getenv("WEBSITE_URL")
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+websiteURL, 302)
	})
	// Link redirect
	r.Get("/{linkID}", func(w http.ResponseWriter, r *http.Request) {
		// Create prepared statements
		selectStatement, err := db.Prepare("SELECT * from links WHERE id = ?")
		if err != nil {
			log.Println("Failed to prepare selectStatement")
			panic(err)
		}
		defer selectStatement.Close()
		updateStatement, err := db.Prepare("UPDATE links SET views=views+1 WHERE id = ?")
		if err != nil {
			log.Println("Failed to prepare updateStatement")
			panic(err)
		}
		defer updateStatement.Close()
		// Get linkID out of URL
		linkID := chi.URLParam(r, "linkID")

		// Convert back to a number
		parsedID, err := base62.FromB62(linkID)

		// See if there was an error while converting
		if err != nil {
			fmt.Fprintf(w, "Invalid link ID format")
		}

		// Now get the URL that this links to
		var rowID int64
		var link string
		var views int64
		err = selectStatement.QueryRow(parsedID).Scan(&rowID, &link, &views)
		if err != nil {
			log.Println("Failed to select link")
			fmt.Fprintf(w, "That link doesn't exist")
			return
		}

		result, err3 := updateStatement.Exec(parsedID)
		if err3 != nil {
			log.Println("Failed to update views")
			http.Redirect(w, r, link, 302)
			return
		}
		rowsAffected, err4 := result.RowsAffected()
		if err4 != nil {
			log.Println("Failed to get rows affected")
			http.Redirect(w, r, link, 302)
			return
		}
		if rowsAffected != 1 {
			log.Println("Rows affected wasn't 1, it was ", rowsAffected)
		}
		http.Redirect(w, r, link, 302)
	})

	return r
}
