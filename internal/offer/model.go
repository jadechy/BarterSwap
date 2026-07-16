package offer

type Offer struct {
	ID           int    `json:"id"`
	ProviderID   int    `json:"provider_id"`
	Titre        string `json:"titre"`
	Description  string `json:"description,omitempty"`
	Categorie    string `json:"categorie"`
	DureeMinutes int    `json:"duree_minutes"`
	Credits      int    `json:"credits"`
	Ville        string `json:"ville,omitempty"`
	Actif        bool   `json:"actif"`
	CreatedAt    string `json:"created_at"`
}

type ListFilter struct {
	Categorie string
	Ville     string
	Search    string
}

var CategoriesValides = []string{
	"Informatique", "Jardinage", "Bricolage", "Cuisine",
	"Musique", "Langues", "Sport", "Tutorat",
	"Déménagement", "Photographie", "Animalier", "Couture", "Autre",
}
