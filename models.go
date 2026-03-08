// models.go
package main

//
//  RÈGLES DE CE FICHIER
// - Aucun calcul
// - Aucun accès réseau
// - Uniquement des structures de données
// - Aligné STRICTEMENT avec le schéma SQL
//

// ==================================================
//  DÉTAILS DU SCORE PORA
// ==================================================
type ScoreDetails struct {
	Followers   float64 `json:"followers"`
	Engagement  float64 `json:"engagement"`
	Orientation float64 `json:"orientation"`
	Cross       float64 `json:"cross"`
}

// ==================================================
//  UNIVERSITÉ
// Table : universites
// Score utilisé : PORA
// ==================================================
type Universite struct {
	ID            string `json:"id"`
	Nom           string `json:"nom"`
	Description   string `json:"description,omitempty"`
	Contacts      string `json:"contacts,omitempty"`
	Email         string `json:"email,omitempty"`
	LienSite      string `json:"lien_site,omitempty"`
	LogoURL       string `json:"logo_url,omitempty"`
	CouvertureURL string `json:"couverture_url,omitempty"`
	Domaine       string `json:"domaine,omitempty"`
	BDEID         string `json:"bde_id,omitempty"`
	Statut        string `json:"statut"`

	ScorePora     float64 `json:"score_pora"`
	ScorePoraPrev float64 `json:"score_pora_prev"` // ✅ AJOUT

	ScoreDetails ScoreDetails `json:"score_details"`
	VideoURL     string       `json:"video_url,omitempty"`
	CreatedAt    string       `json:"created_at"`
}

// ==================================================
//  FOLLOWERS UNIVERSITÉ
// Table : followers_universites
// Signal binaire PORA (popularité)
// ==================================================
type UniversiteFollower struct {
	UserID       string `json:"user_id"`
	UniversiteID string `json:"universite_id"`
	DateFollow   string `json:"date_follow"`
}

// ==================================================
//  ENGAGEMENT UNIVERSITÉ
// Table : engagements_universites
// Signal fort PORA (activité)
// ==================================================
type UniversiteEngagement struct {
	ID           string `json:"id"`
	UniversiteID string `json:"universite_id"`
	Type         string `json:"type"` // like, view, comment, share
	UserID       string `json:"user_id,omitempty"`
	PostID       string `json:"post_id,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// ==================================================
//  CENTRE DE FORMATION
// Table : centres_formation
// Même logique PORA que les universités
// ==================================================
type CentreFormation struct {
	ID            string `json:"id"`
	Nom           string `json:"nom"`
	Description   string `json:"description,omitempty"`
	Contacts      string `json:"contacts,omitempty"`
	Email         string `json:"email,omitempty"`
	LienSite      string `json:"lien_site,omitempty"`
	LogoURL       string `json:"logo_url,omitempty"`
	CouvertureURL string `json:"couverture_url,omitempty"`
	Domaine       string `json:"domaine,omitempty"`
	Statut        string `json:"statut"`

	ScorePora     float64 `json:"score_pora"`
	ScorePoraPrev float64 `json:"score_pora_prev"` // ✅ AJOUT

	ScoreDetails ScoreDetails `json:"score_details"`
	VideoURL     string       `json:"video_url,omitempty"`
	CreatedAt    string       `json:"created_at"`
}

// ==================================================
//  FOLLOWERS CENTRE DE FORMATION
// Table : followers_centres_formation
// ==================================================
type CentreFollower struct {
	UserID     string `json:"user_id"`
	CentreID   string `json:"centre_id"`
	DateFollow string `json:"date_follow"`
}

// ==================================================
//  ENGAGEMENT CENTRE DE FORMATION
// Table : engagements_centres_formation
// ==================================================
type CentreEngagement struct {
	ID        string `json:"id"`
	CentreID  string `json:"centre_id"`
	Type      string `json:"type"` // like, view, comment, share
	UserID    string `json:"user_id,omitempty"`
	PostID    string `json:"post_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

// ==================================================
//  RECOMMANDATIONS CROISÉES
// Table : formation_recommandations_cross
// SIGNAL TRANSVERSAL PORA
// ==================================================
type CrossRecommandation struct {
	FromType string  `json:"from_type"` // universite | centre
	FromID   string  `json:"from_id"`
	ToType   string  `json:"to_type"` // universite | centre
	ToID     string  `json:"to_id"`
	Poids    float64 `json:"poids"`
	Raison   string  `json:"raison,omitempty"`
}

// ==================================================
// ORIENTATION — VECTEUR UTILISATEUR
// ==================================================
type OrientationVector struct {
	UserID     string    `json:"user_id"`
	Profile    []float64 `json:"profile"`    // vecteur normalisé
	Confidence float64   `json:"confidence"` // confiance globale
}

// ==================================================
// ORIENTATION — QUIZ BRUT
// ==================================================
type QuizSubmission struct {
	UserID      string             `json:"user_id"`
	QuizVersion string             `json:"quiz_version"`
	Responses   map[string]float64 `json:"responses"`
}

// ==================================================
// ORIENTATION → RECOMMANDATION (PROA → PORA)
// Table : orientation_recommendations
// ==================================================
type OrientationRecommendation struct {
	ID string `json:"id"`

	UserID    string `json:"user_id"`
	ProfileID string `json:"profile_id"`

	TargetType string `json:"target_type"` // universite | centre
	TargetID   string `json:"target_id"`

	Score      float64 `json:"score"`      // score PROA (0 → 1)
	Rank       int     `json:"rank"`       // position dans le classement PROA
	Confidence float64 `json:"confidence"` // confiance locale
	Raison     string  `json:"raison,omitempty"`

	CreatedAt string `json:"created_at"`
}

type PORANode struct {
	ID   string `json:"id"`
	Type string `json:"type"` // universite | centre

	Nom string `json:"nom,omitempty"` // ✅ AJOUT (enrichissement)

	ScoreRaw float64 `json:"score_raw"`
	Score    float64 `json:"score"`

	Rank       int `json:"rank"`
	Percentile int `json:"percentile"`

	Trend  PORATrend    `json:"trend"`
	Detail ScoreDetails `json:"detail"`
}
