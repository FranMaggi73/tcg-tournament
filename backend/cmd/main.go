package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/FranMaggi73/tcg-tournament/backend/internal/handlers"
	"github.com/FranMaggi73/tcg-tournament/backend/internal/middleware"
	"github.com/FranMaggi73/tcg-tournament/backend/internal/tournament"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	// Load .env file if available
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	ctx := context.Background()

	// Firebase Credentials
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	config := &firebase.Config{
		ProjectID: os.Getenv("FIREBASE_PROJECT_ID"),
	}
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error initializing firestore client: %v", err)
	}
	defer client.Close()

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error initializing firebase auth client: %v", err)
	}

	// Services & Repositories
	repo := tournament.NewRepository(client)
	swiss := tournament.NewSwissService(repo)
	h := handlers.NewTournamentHandler(repo, swiss)
	fh := handlers.NewFriendshipHandler(repo)

	// Router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// Public Routes
	r.GET("/tournaments/:id", h.GetTournament)
	r.GET("/tournaments/:id/standings", h.GetStandings)
	r.GET("/tournaments/:id/export", h.ExportStandings)
	r.POST("/tournaments/:id/players", h.RegisterPlayer)
	r.POST("/tournaments/join", h.JoinTournamentByCode)

	// Protected Routes (Require Auth)
	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware(authClient))
	{
		authGroup.POST("/tournaments", h.CreateTournament)
		authGroup.DELETE("/tournaments/:id", h.DeleteTournament)
		authGroup.PATCH("/tournaments/:id/complete", h.CompleteTournament)
		authGroup.POST("/tournaments/:id/rounds/next", h.NextRound)
		authGroup.PATCH("/tournaments/:id/matches/:matchId", h.UpdateMatchResult)
		authGroup.PATCH("/tournaments/:id/players/:playerId/status", h.UpdatePlayerStatus)
		authGroup.DELETE("/tournaments/:id/players/:playerId", h.RemovePlayer)
		authGroup.POST("/tournaments/:id/rollback", h.RollbackRound)

		// Friendship Routes
		authGroup.POST("/friends", fh.AddFriend)
		authGroup.GET("/friends", fh.GetFriends)
		authGroup.GET("/friends/pending", fh.GetPendingRequests)
		authGroup.PATCH("/friends/:id", fh.UpdateFriendshipStatus)
	}

	// Graceful Shutdown Setup
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Server starting on :8080...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}