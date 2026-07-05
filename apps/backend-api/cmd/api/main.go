package main

import (
	"log"
	"net/http"
	"os"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/backend-api/internal/blog"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/backend-api/internal/db"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/backend-api/internal/handler"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/backend-api/internal/mail"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var pinger handler.Pinger
	var contactStore handler.ContactStore
	var postStore blog.Store
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		pool, err := db.Open(databaseURL)
		if err != nil {
			log.Fatalf("invalid DATABASE_URL: %v", err)
		}
		defer pool.Close()

		if err := db.Migrate(pool); err != nil {
			log.Fatalf("migrations: %v", err)
		}
		log.Print("database migrations applied")

		pinger = pool
		contactStore = db.ContactStore{DB: pool}
		postStore = db.PostStore{DB: pool}
	} else {
		log.Print("DATABASE_URL not set; /readyz will not check the database")
	}

	var mailer handler.Mailer
	if smtp := mail.FromEnv(); smtp != nil {
		mailer = smtp
		log.Print("SMTP configured; contact messages will also notify by email")
	} else {
		log.Print("SMTP not configured; contact messages will only be stored")
	}

	blogToken := os.Getenv("BLOG_ADMIN_TOKEN")
	if blogToken == "" {
		log.Print("BLOG_ADMIN_TOKEN not set; POST /posts is disabled")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Health)
	mux.Handle("GET /readyz", handler.Ready(pinger))
	mux.Handle("POST /contact", handler.Contact(contactStore, mailer))
	mux.Handle("GET /posts", handler.ListPosts(postStore))
	mux.Handle("GET /posts/{slug}", handler.GetPost(postStore))
	mux.Handle("POST /posts", handler.CreatePost(postStore, blogToken))

	log.Printf("backend-api listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
