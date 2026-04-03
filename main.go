// *********** Main.go **************
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// ------------------------------------------------
	// Chargement configuration globale
	// (le .env est déjà géré dans LoadConfig)
	// ------------------------------------------------
	LoadConfig()

	// ------------------------------------------------
	// Mode Gin (debug / release)
	// ------------------------------------------------
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	log.Println("🚀 PORA engine démarré")

	// ------------------------------------------------
	// Initialisation serveur HTTP
	// ------------------------------------------------
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 🔐 Sécurité proxies (supprime le warning)
	r.SetTrustedProxies(nil)

	// ------------------------------------------------
	// Middleware CORS pour autoriser le front local
	// ------------------------------------------------
	r.Use(cors.New(cors.Config{
		// Allow common local dev origins (add other frontend ports as needed)
		AllowOrigins: []string{
			"http://127.0.0.1:5500",
			"http://127.0.0.1:5501",
			"http://127.0.0.1:5502",
			"http://localhost:5500",
			"http://localhost:5501",
			"http://localhost:5502",
			"http://127.0.0.1:8000",
			"http://localhost:8000",
			"https://pora-frontend.onrender.com",
			"https://universearch-pora-frontend.onrender.com",
		}, // front local
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "x-user-id", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ------------------------------------------------
	// Endpoint de santé
	// ------------------------------------------------
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "pora-engine",
		})
	})

	// ------------------------------------------------
	// Enregistrement des routes métier
	// ------------------------------------------------
	RegisterRoutes(r)

	// ------------------------------------------------
	// Lancement du cron automatique PORA
	// ------------------------------------------------
	StartCron()

	// ------------------------------------------------
	// Démarrage du serveur
	// ------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("📡 API PORA disponible sur :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
