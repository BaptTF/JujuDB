# JujuDB - Gestionnaire d'Inventaire Familial 🧊

JujuDB est une application web familiale pour gérer l'inventaire de vos espaces de stockage (congélateur, réfrigérateur, garde-manger, etc.).

## Fonctionnalités

- 🔐 **Authentification simple** : Un seul mot de passe familial avec cookies longue durée
- 🔍 **Recherche avancée** : Recherche intelligente avec distance de Levenshtein
- 🏷️ **Filtres multiples** : Filtrage par emplacement et catégorie
- 📱 **Interface responsive** : Optimisée pour mobile et desktop
- 🐳 **Déploiement Docker** : Configuration complète avec PostgreSQL
- 🇫🇷 **Interface en français** : Entièrement localisée

## Installation et Démarrage

### Prérequis
- Docker et Docker Compose
- Git

### Démarrage rapide

1. **Cloner le projet**
   ```bash
   git clone <votre-repo>
   cd JujuDB
   ```

2. **Configurer les variables d'environnement** (optionnel)
   
   Vous pouvez modifier les variables dans `docker-compose.yml` :
   - `APP_PASSWORD` : Mot de passe de connexion (défaut: `your-secure-app-password`)
   - `SESSION_KEY` : Clé de session (changez en production)
   - Paramètres de base de données

3. **Démarrer l'application**
   ```bash
   docker-compose up -d
   ```

4. **Accéder à l'application**
   
   Ouvrez votre navigateur sur `http://localhost:8080`
   
   Mot de passe par défaut : `your-secure-app-password`

### Arrêter l'application

```bash
docker-compose down
```

### Voir les logs

```bash
docker-compose logs -f jujudb
```

## Utilisation

### Connexion
- Utilisez le mot de passe familial configuré (défaut: `your-secure-app-password`)
- Le cookie de session dure 30 jours

### Gestion des articles
- **Ajouter** : Cliquez sur "+ Ajouter un article"
- **Modifier** : Cliquez sur "Modifier" sur une carte d'article
- **Supprimer** : Cliquez sur "Supprimer" (avec confirmation)

### Recherche et filtres
- **Recherche textuelle** : Tapez dans la barre de recherche
- **Filtres** : Utilisez les menus déroulants pour filtrer par emplacement/catégorie
- **Recherche avancée** : Utilise l'algorithme de Levenshtein pour des résultats intelligents

### Emplacements supportés
- Congélateur
- Réfrigérateur  
- Garde-manger
- Cave
- Garage

### Catégories supportées
- Viande
- Poisson
- Légumes
- Fruits
- Plats préparés
- Desserts
- Autres

## Architecture Technique

### Backend (Go)
- **Framework** : Gorilla Mux pour le routage
- **Base de données** : PostgreSQL avec driver `lib/pq`
- **Sessions** : Gorilla Sessions avec cookies sécurisés
- **Recherche** : Algorithme de Levenshtein pour la recherche floue
- **API REST** : Endpoints JSON pour toutes les opérations CRUD

### Frontend
- **HTML/CSS/JavaScript** : Interface responsive moderne
- **Design** : Interface utilisateur intuitive avec animations
- **AJAX** : Communication asynchrone avec l'API

### Base de données
- **PostgreSQL** : Base de données relationnelle
- **Indexes** : Optimisation des requêtes sur nom, emplacement, catégorie
- **Données d'exemple** : Jeu de données initial pour les tests

### Docker
- **Multi-stage build** : Image optimisée pour la production
- **Health checks** : Vérification de l'état des services
- **Volumes persistants** : Sauvegarde des données PostgreSQL

## Configuration

### Variables d'environnement

| Variable | Description | Défaut |
|----------|-------------|---------|
| `PORT` | Port d'écoute de l'application | `8080` |
| `DB_HOST` | Hôte PostgreSQL | `postgres` |
| `DB_USER` | Utilisateur PostgreSQL | `jujudb` |
| `DB_PASSWORD` | Mot de passe PostgreSQL | `your-secure-postgres-password` |
| `DB_NAME` | Nom de la base de données | `jujudb` |
| `SESSION_KEY` | Clé de chiffrement des sessions | (à changer) |
| `APP_PASSWORD` | Mot de passe de l'application | `your-secure-app-password` |

### Sécurité

⚠️ **Important pour la production** :
- Changez `SESSION_KEY` par une clé aléatoire forte
- Modifiez `APP_PASSWORD` 
- Utilisez HTTPS en production
- Configurez des mots de passe PostgreSQL forts

## Développement

### Développement local sans Docker

1. **Installer PostgreSQL** localement
2. **Créer la base de données**
   ```sql
   CREATE DATABASE jujudb;
   CREATE USER jujudb WITH PASSWORD 'your-secure-postgres-password';
   GRANT ALL PRIVILEGES ON DATABASE jujudb TO jujudb;
   ```
3. **Installer les dépendances Go**
   ```bash
   go mod download
   ```
4. **Démarrer l'application**
   ```bash
   go run main.go
   ```

### Structure du projet

```
JujuDB/
├── main.go              # Application principale Go
├── go.mod               # Dépendances Go
├── go.sum               # Checksums des dépendances
├── Dockerfile           # Configuration Docker
├── docker-compose.yml   # Orchestration des services
├── init.sql            # Script d'initialisation DB
├── templates/          # Templates HTML
│   ├── login.html
│   └── dashboard.html
└── static/             # Ressources statiques
    ├── css/
    │   └── style.css
    └── js/
        └── app.js
```

## API Endpoints

### Authentification
- `GET /` - Redirection vers login ou dashboard
- `GET /login` - Page de connexion
- `POST /login` - Authentification
- `POST /logout` - Déconnexion

### Interface
- `GET /dashboard` - Interface principale (authentification requise)

### API REST (authentification requise)
- `GET /api/items` - Liste des articles (avec filtres optionnels)
- `POST /api/items` - Créer un article
- `PUT /api/items/{id}` - Modifier un article
- `DELETE /api/items/{id}` - Supprimer un article
- `GET /api/search?q={query}` - Recherche avec Levenshtein

## Extensibilité

L'application est conçue pour être facilement étendue :
- **Nouveaux emplacements** : Ajoutez-les dans les options HTML et CSS
- **Nouvelles catégories** : Modifiez les listes déroulantes
- **Champs supplémentaires** : Étendez la structure `Item` et les formulaires
- **Authentification avancée** : Remplacez le système de mot de passe unique

## Support et Contribution

Cette application est conçue pour un usage familial. Pour des questions ou améliorations :
1. Consultez les logs avec `docker-compose logs`
2. Vérifiez la configuration des variables d'environnement
3. Assurez-vous que PostgreSQL est accessible

## Licence

Projet familial - Usage libre