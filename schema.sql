DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS credit_transactions;
DROP TABLE IF EXISTS exchanges;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    pseudo VARCHAR(100) NOT NULL UNIQUE,
    bio TEXT,
    ville VARCHAR(100),
    credit_balance INT NOT NULL DEFAULT 10,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE skills (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    nom VARCHAR(100) NOT NULL,
    niveau ENUM('débutant', 'intermédiaire', 'expert') NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE services (
    id INT AUTO_INCREMENT PRIMARY KEY,
    provider_id INT NOT NULL,
    titre VARCHAR(255) NOT NULL,
    description TEXT,
    categorie ENUM(
        'Informatique', 'Jardinage', 'Bricolage', 'Cuisine',
        'Musique', 'Langues', 'Sport', 'Tutorat',
        'Déménagement', 'Photographie', 'Animalier', 'Couture', 'Autre'
    ) NOT NULL,
    duree_minutes INT NOT NULL,
    credits INT NOT NULL,
    ville VARCHAR(100),
    actif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (provider_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE exchanges (
    id INT AUTO_INCREMENT PRIMARY KEY,
    service_id INT NOT NULL,
    requester_id INT NOT NULL,
    owner_id INT NOT NULL,
    status ENUM('pending', 'accepted', 'rejected', 'cancelled', 'completed') NOT NULL DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(id),
    FOREIGN KEY (requester_id) REFERENCES users(id),
    FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE TABLE credit_transactions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    exchange_id INT,
    montant INT NOT NULL,
    type ENUM('earn', 'spend', 'refund', 'welcome') NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (exchange_id) REFERENCES exchanges(id)
);

CREATE TABLE reviews (
    id INT AUTO_INCREMENT PRIMARY KEY,
    exchange_id INT NOT NULL,
    author_id INT NOT NULL,
    target_id INT NOT NULL,
    note INT NOT NULL CHECK (note BETWEEN 1 AND 5),
    commentaire TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_review(exchange_id, author_id),
    FOREIGN KEY (exchange_id) REFERENCES exchanges(id),
    FOREIGN KEY (author_id) REFERENCES users(id),
    FOREIGN KEY (target_id) REFERENCES users(id)
);