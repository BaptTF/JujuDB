# JujuDB - Guide de Déploiement Production

## Configuration Traefik pour jujudb.bapttf.com

### Prérequis
- Serveur avec Docker et Docker Compose
- Traefik configuré avec réseau externe `traefik`
- Certificats SSL Let's Encrypt configurés
- DNS pointant vers votre serveur

### Déploiement Production

1. **Cloner le projet sur le serveur**
   ```bash
   git clone <votre-repo> /opt/jujudb
   cd /opt/jujudb
   ```

2. **Configurer les variables d'environnement**
   
   Copiez le fichier d'exemple et configurez vos valeurs :
   ```bash
   cp .env.prod .env
   # Éditez .env avec vos vraies valeurs
   nano .env
   ```
   
   Ou créez directement le fichier `.env` :
   ```bash
   # Base de données
   POSTGRES_DB=jujudb
   POSTGRES_USER=jujudb
   POSTGRES_PASSWORD=your-secure-postgres-password
   
   # Application
   APP_PASSWORD=your-secure-app-password
   SESSION_KEY=your-super-secret-session-key-32-characters-minimum
   
   # Production
   PRODUCTION=true
   HTTPS=true
   
   # Database Connection
   DB_HOST=postgres
   DB_USER=jujudb
   DB_PASSWORD=your-secure-postgres-password
   DB_NAME=jujudb
   PORT=8080
   ```

3. **Démarrer l'application**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

4. **Vérifier le déploiement**
   ```bash
   docker-compose -f docker-compose.prod.yml logs -f jujudb
   ```

### Configuration Traefik

L'application est configurée avec les labels Traefik suivants :

- **Domaine** : `jujudb.bapttf.com`
- **HTTPS** : Certificats Let's Encrypt automatiques
- **Sécurité** : Headers de sécurité HTTPS
- **Réseau** : Connecté au réseau Traefik externe

### Sécurité Production

✅ **Configuré automatiquement** :
- Cookies sécurisés HTTPS
- Headers de sécurité Traefik
- Redirection HTTP → HTTPS
- HSTS (HTTP Strict Transport Security)
- Isolation réseau Docker

⚠️ **À configurer manuellement** :
- Changez `SESSION_KEY` par une clé aléatoire forte (32+ caractères)
- Configurez des sauvegardes PostgreSQL
- Surveillez les logs d'accès

### Commandes de Gestion

```bash
# Démarrer
docker-compose -f docker-compose.prod.yml up -d

# Arrêter
docker-compose -f docker-compose.prod.yml down

# Voir les logs
docker-compose -f docker-compose.prod.yml logs -f

# Redémarrer après mise à jour
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml build --no-cache
docker-compose -f docker-compose.prod.yml up -d

# Sauvegarde base de données
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U jujudb jujudb > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Mise à Jour

1. **Arrêter l'application**
   ```bash
   docker-compose -f docker-compose.prod.yml down
   ```

2. **Mettre à jour le code**
   ```bash
   git pull origin main
   ```

3. **Reconstruire et redémarrer**
   ```bash
   docker-compose -f docker-compose.prod.yml build --no-cache
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Surveillance

- **URL** : https://jujudb.bapttf.com
- **Mot de passe** : `your-secure-app-password`
- **Logs** : `docker-compose -f docker-compose.prod.yml logs -f`
- **Base de données** : PostgreSQL avec volume persistant

### Résolution de Problèmes

1. **Vérifier que Traefik fonctionne**
   ```bash
   docker network ls | grep traefik
   ```

2. **Vérifier les certificats SSL**
   ```bash
   curl -I https://jujudb.bapttf.com
   ```

3. **Vérifier la connectivité base de données**
   ```bash
   docker-compose -f docker-compose.prod.yml exec jujudb nc -z postgres 5432
   ```

4. **Accéder à la base de données**
   ```bash
   docker-compose -f docker-compose.prod.yml exec postgres psql -U jujudb -d jujudb
   ```
