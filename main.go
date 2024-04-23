package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"follower.xws.com/handler"
	"follower.xws.com/repo"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// //Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[follower-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[follower-store] ", log.LstdFlags)

	// NoSQL: Initialize Movie Repository store
	store, err := repo.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseDriverConnection(timeoutContext)
	store.CheckConnection()
	//cartHandler := &handler.ShoppingCartHandler{ShoppingCartService: cartService}
	FollowersHandler := handler.NewFollowersHandler(logger, store)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/createFollower", FollowersHandler.CreateFollowing).Methods("POST")

	permittedHeaders := handlers.AllowedHeaders([]string{"Requested-With", "Content-Type", "Authorization"})
	permittedOrigins := handlers.AllowedOrigins([]string{"*"})
	permittedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})

	// Start server
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
	println("Server starting")
	log.Fatal(http.ListenAndServe(":8084", handlers.CORS(permittedHeaders, permittedOrigins, permittedMethods)(router)))
}
