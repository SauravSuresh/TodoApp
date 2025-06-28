package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/SauravSuresh/todoapp/common"
	"github.com/SauravSuresh/todoapp/handlers"
	"github.com/SauravSuresh/todoapp/middlewares"
	"github.com/SauravSuresh/todoapp/utils"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TodoHandlers() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middlewares.AuthenticationMiddelware)
	router.Use(middlewares.UserLoaderMiddleware)
	router.Group(func(r chi.Router) {
		r.Get("/index", handlers.IndexHandler)
		r.Get("/", handlers.GetTodoHandler)
		r.Post("/", handlers.CreateTodoHandler)
		r.Put("/{id}", handlers.UpdateTodoHandler)
		r.Delete("/{id}", handlers.DeleteTodoHandler)
		r.Get("/createdbyme", handlers.GetCreatedTodoHandler)
		r.Get("/assignedtome", handlers.GetAssignedTodoHandler)
		r.Post("/setstatus/{id}", handlers.SetStatusHandler)
	})
	return router
}

func LoginHandlers() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Group(func(r chi.Router) {
		r.Post("/register", handlers.RegisterUserHandler)
		r.Get("/register", handlers.HomeHandler)
		r.Get("/login", handlers.LoginPageHandler)
		r.Post("/login", handlers.LoginAttemptHandler)
		r.Post("/logout", handlers.Logout)
		r.Get("/users", handlers.GetAvaialableUsers)
	})
	return router
}

func init() {
	fmt.Println("init function called")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var err error
	common.Client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	utils.CheckErr(err, "failed to connect to db")
	if err != nil {
		return
	}

	// Use the correct function to get the database name
	common.Db = common.Client.Database(common.GetDbName())
}

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	fs := http.FileServer(http.Dir("./static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	router.Get("/", handlers.HomeHandler)

	router.Mount("/auth", LoginHandlers())
	router.Mount("/todo", TodoHandlers())

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	server := http.Server{
		Addr:         ":9000",
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Handler:      router,
	}

	go func() {
		fmt.Printf("Serving on port: %v", 9000)
		err := server.ListenAndServe()
		utils.CheckErr(err, "Error starting server: ")

	}()
	sig := <-stopChan

	log.Printf("Interrupt signal recieved %v\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {

		log.Fatalf("Server shutdown failed: %v\n", err)

	}

	log.Printf("server shutdown gracefully")

}
