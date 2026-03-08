// Cron.go
package main

import (
	"log"
	"time"
)

/*
Ce cron déclenche automatiquement le recalcul du score PORA.

Il n’exécute aucune intelligence :
- il lance simplement l’agrégation des signaux utilisateurs
- followers
- engagements
- recommandations d’orientation (Python)

Le calcul est périodique et déterministe.
*/

func StartCron() {
	go func() {
		log.Println("[cron][PORA] Lancement initial du classement")

		if err := RunRanking(); err != nil {
			log.Println("[cron][PORA] erreur universites:", err)
		}
		if err := RunRankingCentres(); err != nil {
			log.Println("[cron][PORA] erreur centres:", err)
		}

		ticker := time.NewTicker(time.Duration(CronIntervalHours) * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("[cron][PORA] Recalcul périodique")

			if err := RunRanking(); err != nil {
				log.Println("[cron][PORA] erreur universites:", err)
			}
			if err := RunRankingCentres(); err != nil {
				log.Println("[cron][PORA] erreur centres:", err)
			}
		}
	}()
}
