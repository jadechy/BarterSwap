SET FOREIGN_KEY_CHECKS = 0;
TRUNCATE TABLE reviews;
TRUNCATE TABLE credit_transactions;
TRUNCATE TABLE exchanges;
TRUNCATE TABLE services;
TRUNCATE TABLE skills;
TRUNCATE TABLE users;
SET FOREIGN_KEY_CHECKS = 1;

-- Utilisateurs
INSERT INTO users (id, pseudo, bio, ville, credit_balance) VALUES
(1, 'jade', 'Développeuse web passionnée', 'Paris', 15),
(2, 'thomas', 'Musicien et bricoleur du dimanche', 'Lyon', 8),
(3, 'camille', 'Photographe freelance', 'Bordeaux', 20),
(4, 'lucas', 'Jardinier amateur et cuisinier', 'Paris', 5);

-- Compétences
INSERT INTO skills (user_id, nom, niveau) VALUES
(1, 'Informatique', 'expert'),
(1, 'Tutorat', 'intermediaire'),
(2, 'Musique', 'expert'),
(2, 'Bricolage', 'intermediaire'),
(3, 'Photographie', 'expert'),
(3, 'Informatique', 'debutant'),
(4, 'Jardinage', 'expert'),
(4, 'Cuisine', 'intermediaire');

-- Services
INSERT INTO services (id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif) VALUES
(1, 1, 'Cours d\'initiation au Go', 'Apprentissage des bases du langage Go', 'Informatique', 60, 2, 'Paris', TRUE),
(2, 1, 'Aide à la création de site web', 'Conception et développement d\'un site vitrine', 'Informatique', 120, 4, 'Paris', TRUE),
(3, 2, 'Cours de guitare débutant', 'Initiation à la guitare acoustique', 'Musique', 60, 2, 'Lyon', TRUE),
(4, 2, 'Réparation de meubles', 'Remise en état de meubles abîmés', 'Bricolage', 90, 3, 'Lyon', TRUE),
(5, 3, 'Séance photo portrait', 'Shooting photo professionnel en extérieur', 'Photographie', 120, 5, 'Bordeaux', TRUE),
(6, 4, 'Entretien de jardin', 'Taille, désherbage et plantation', 'Jardinage', 180, 3, 'Paris', TRUE),
(7, 4, 'Cours de cuisine italienne', 'Apprentissage de recettes italiennes traditionnelles', 'Cuisine', 90, 2, 'Paris', TRUE);

-- Échanges
INSERT INTO exchanges (id, service_id, requester_id, owner_id, status) VALUES
(1, 3, 1, 2, 'completed'),
(2, 1, 3, 1, 'completed'),
(3, 5, 4, 3, 'accepted'),
(4, 6, 2, 4, 'pending'),
(5, 1, 4, 1, 'rejected');

-- Transactions de crédits
INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES
(1, NULL, 10, 'welcome'),
(2, NULL, 10, 'welcome'),
(3, NULL, 10, 'welcome'),
(4, NULL, 10, 'welcome'),
(1, 1, -2, 'spend'),
(2, 1, 2, 'earn'),
(3, 2, -2, 'spend'),
(1, 2, 2, 'earn'),
(4, 3, -5, 'spend');

-- Avis sur les échanges terminés
INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire) VALUES
(1, 1, 2, 5, 'Super cours, très pédagogue !'),
(1, 2, 1, 4, 'Élève sérieuse et motivée'),
(2, 3, 1, 5, 'Excellente aide, très claire dans ses explications'),
(2, 1, 3, 4, 'Très agréable à aider');