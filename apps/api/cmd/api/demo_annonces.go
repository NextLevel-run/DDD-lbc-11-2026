package main

import (
	"time"

	"ddd-second-hand-marketplace/internal/bounded-context/annonce/domain"
)

// motDePasseDemo is the mot de passe of every demo annonce.
const motDePasseDemo = "demo1234"

// loadDemoAnnonces fills the repository with annonces to look at in
// development, varied in categorie, prix and region, and returns their IDs.
// Tests never load them.
func loadDemoAnnonces(repo domain.AnnonceRepository) ([]string, error) {
	now := time.Now()
	ago := func(hours int) time.Time { return now.Add(-time.Duration(hours) * time.Hour) }

	demo := []domain.NewAnnonceArgs{
		{
			Titre: "Peugeot 208 essence, 45 000 km", Prix: 1250000, Categorie: "vehicules",
			Description: "Première main, carnet d'entretien à jour, contrôle technique OK. Pneus neufs.",
			Ville:       "Lyon", CodePostal: "69003", Pseudo: "lyonnais69", Email: "lyonnais69@example.com",
			Telephone: "0611223344", DeposeeLe: ago(2),
		},
		{
			Titre: "Studio meublé proche gare", Prix: 9500000, Categorie: "immobilier",
			Description: "Studio de 22 m² entièrement rénové, idéal investissement locatif.",
			Ville:       "Rennes", CodePostal: "35000", Pseudo: "breizh_immo", Email: "breizh@example.com",
			DeposeeLe: ago(5),
		},
		{
			Titre: "iPhone 13 128 Go", Prix: 38000, Categorie: "multimedia",
			Description: "Très bon état, batterie à 89 %, vendu avec sa coque et son chargeur.",
			Ville:       "Paris", CodePostal: "75015", Pseudo: "parisien", Email: "parisien@example.com",
			Telephone: "0700000001", DeposeeLe: ago(9),
		},
		{
			Titre: "Canapé trois places en velours", Prix: 15000, Categorie: "maison-jardin",
			Description: "Canapé vert sapin, quelques traces d'usure, à venir chercher sur place.",
			Ville:       "Nantes", CodePostal: "44000", Pseudo: "nantaise", Email: "nantaise@example.com",
			DeposeeLe: ago(20),
		},
		{
			Titre: "Vélo électrique de ville", Prix: 89000, Categorie: "loisirs",
			Description: "Autonomie 60 km, révisé il y a un mois. Antivol fourni.",
			Ville:       "Saint-Étienne", CodePostal: "42000", Pseudo: "stephanois", Email: "stephanois@example.com",
			DeposeeLe: ago(30),
		},
		{
			Titre: "Veste en jean taille M", Prix: 1500, Categorie: "mode",
			Description: "Portée deux fois, comme neuve. Envoi possible.",
			Ville:       "Bordeaux", CodePostal: "33000", Pseudo: "bordelaise", Email: "bordelaise@example.com",
			DeposeeLe: ago(48),
		},
		{
			Titre: "Cours de guitare à domicile", Prix: 2500, Categorie: "emploi-services",
			Description: "Professeur diplômé, tous niveaux, 25 € l'heure.",
			Ville:       "Marseille", CodePostal: "13001", Pseudo: "guitariste13", Email: "guitare@example.com",
			Telephone: "0699887766", DeposeeLe: ago(72),
		},
		{
			Titre: "Cartons de déménagement à donner", Prix: 0, Categorie: "autres",
			Description: "Une vingtaine de cartons solides, à récupérer rapidement.",
			Ville:       "Lille", CodePostal: "59000", Pseudo: "lillois", Email: "lillois@example.com",
			DeposeeLe: ago(96),
		},
		{
			Titre: "Scooter 50 cm³", Prix: 70000, Categorie: "vehicules",
			Description: "Idéal pour la ville, entretien régulier, deux casques offerts.",
			Ville:       "Ajaccio", CodePostal: "20000", Pseudo: "corsica", Email: "corsica@example.com",
			DeposeeLe: ago(120),
		},
		{
			Titre: "Planche de surf 7'2", Prix: 22000, Categorie: "loisirs",
			Description: "Planche évolutive en bon état, dérives incluses.",
			Ville:       "Saint-Denis", CodePostal: "97400", Pseudo: "surfeur974", Email: "surf974@example.com",
			DeposeeLe: ago(150),
		},
		{
			Titre: "Table de jardin en teck", Prix: 18000, Categorie: "maison-jardin",
			Description: "Table pour six personnes, huilée chaque printemps.",
			Ville:       "Saint-Denis", CodePostal: "93200", Pseudo: "dionysien", Email: "dionysien@example.com",
			DeposeeLe: ago(200),
		},
	}

	for i := range demo {
		demo[i].MotDePasse = motDePasseDemo
	}

	ids := make([]string, 0, len(demo))
	for _, args := range demo {
		annonce, err := domain.NewAnnonce(args)
		if err != nil {
			return nil, err
		}
		// Demo annonces skip the moderation: they are online from the start.
		if err := annonce.Activer(); err != nil {
			return nil, err
		}
		if err := repo.Save(annonce); err != nil {
			return nil, err
		}
		ids = append(ids, annonce.ID())
	}
	return ids, nil
}
