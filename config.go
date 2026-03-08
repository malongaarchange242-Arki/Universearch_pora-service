// Config.go
package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (

	// Supabase

	SupabaseURL     string
	SupabaseAnon    string
	SupabaseService string

	// Orientation (microservice Python)

	OrientationServiceURL string

	// Cron PORA

	CronIntervalHours int

	// Runtime

	DebugMode bool

	configLoaded = false
)

// Chargement de la configuration globale (idempotent)

func LoadConfig() {
	if configLoaded {
		return
	}
	configLoaded = true

	// CHARGEMENT DU .env (DEV uniquement)

	if err := godotenv.Load(); err != nil {
		log.Println(" Aucun fichier .env trouvé (variables système utilisées)")
	}

	// Helper variable obligatoire (fail fast)

	getMandatory := func(key string) string {
		val := strings.TrimSpace(os.Getenv(key))
		if val == "" {
			log.Fatalf(" CONFIG ERROR: la variable %s est manquante", key)
		}
		return val
	}

	// VARIABLES OBLIGATOIRES

	SupabaseURL = getMandatory("SUPABASE_URL")
	SupabaseService = getMandatory("SUPABASE_SERVICE_KEY")

	// VARIABLES OPTIONNELLES

	SupabaseAnon = strings.TrimSpace(os.Getenv("SUPABASE_ANON_KEY"))
	OrientationServiceURL = strings.TrimSpace(os.Getenv("ORIENTATION_SERVICE_URL"))

	// VALIDATIONS

	if !strings.HasPrefix(SupabaseURL, "https://") {
		log.Fatalf(
			" SUPABASE_URL doit commencer par https:// → reçu: %s",
			SupabaseURL,
		)
	}

	if OrientationServiceURL != "" &&
		!strings.HasPrefix(OrientationServiceURL, "http://") &&
		!strings.HasPrefix(OrientationServiceURL, "https://") {
		log.Fatalf(
			" ORIENTATION_SERVICE_URL invalide (http/https requis) : %s",
			OrientationServiceURL,
		)
	}

	// CRON PORA

	CronIntervalHours = getIntEnv("PORA_CRON_INTERVAL_HOURS", 24)
	if CronIntervalHours <= 0 {
		log.Fatalf(
			" PORA_CRON_INTERVAL_HOURS doit être > 0 → reçu: %d",
			CronIntervalHours,
		)
	}

	// DEBUG

	DebugMode = getBoolEnv("DEBUG", false)

	log.Println(" Configuration chargée : Supabase + PORA + Orientation + Cron")
}

// HELPERS

func getIntEnv(key string, defaultVal int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf(" %s doit être un entier : %v", key, err)
		}
		return n
	}
	return defaultVal
}

func getBoolEnv(key string, defaultVal bool) bool {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			log.Fatalf(" %s doit être un booléen : %v", key, err)
		}
		return b
	}
	return defaultVal
}
