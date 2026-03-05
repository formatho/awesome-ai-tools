package main

import (
	"log"
	"os"

	"github.com/formatho/agent-orchestrator/backend/internal/api"
	"github.com/formatho/agent-orchestrator/backend/store"
)

func main() {
	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/agent-orchestrator.db"
	}

	db, err := store.InitDB(dbPath)
	if err != nil {
		log.Fatal("Failed to init database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := store.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Create API server
	server := api.NewServer(db)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = ":18765"
	}

	log.Printf("🚀 Server starting on http://localhost%s", port)
	log.Printf("📊 API: http://localhost%s/api", port)
	log.Printf("🔌 WebSocket: ws://localhost%s/ws", port)

	if err := server.Listen(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
