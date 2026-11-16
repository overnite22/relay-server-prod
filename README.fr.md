[English](README.md) · **Français**

# Relay

Gestionnaire de connexions à distance auto-hébergé, avec un ruban de style RDM, une arborescence de dossiers et des onglets de session.
Les consoles SSH, RDP et Proxmox s'exécutent dans le navigateur. Un coffre d'identifiants chiffré, le transfert de fichiers,
la synchronisation d'inventaire, les notes et les enregistrements sont inclus.

**Les comptes locaux fonctionnent immédiatement. Le SSO est optionnel.** Relay s'adresse à un homelab de confiance ou à une équipe
d'administration qui partage un inventaire et un coffre. Il n'isole pas des locataires qui ne se font pas confiance.

## Installation avec Docker Compose

Nécessite Docker Engine et Compose v2. Aucun fournisseur d'identité ni compilation de l'application n'est requis.

1. Téléchargez et extrayez le ZIP d'installation Compose d'une version publiée (tag). Il contient
   `compose.yaml` et `.env.example`.
2. Copiez `.env.example` vers `.env`, puis exécutez :

   ```sh
   docker compose pull
   docker compose up -d
   docker compose exec relay /relay setup-token
   ```

3. Ouvrez `http://localhost:8080` sur l'hôte Docker. Saisissez le code à usage unique et créez le premier
   administrateur avec un mot de passe d'au moins 12 caractères. Le fichier du code est supprimé après la configuration.
4. Administration → **Comptes locaux** permet de créer des utilisateurs, d'attribuer le rôle d'administrateur, de réinitialiser
   les mots de passe et de désactiver des comptes. Les utilisateurs peuvent changer leur propre mot de passe et se déconnecter depuis la barre de titre.

Pour un hôte Docker distant, utilisez une redirection de port SSH pour l'accès initial, ou configurez un
reverse proxy HTTPS. Définissez `RELAY_PUBLIC_ORIGIN=https://relay.example.com` et transmettez les WebSockets,
`Host` et `X-Forwarded-Proto` vers `127.0.0.1:8080`. Le port publié est lié à la boucle locale par défaut.
Les cookies de session HTTP ne sont autorisés que pour une origine de boucle locale configurée. Utilisez HTTPS pour l'accès distant.

Les volumes nommés sont initialisés avec les droits requis. Le volume de données contient la clé maîtresse du coffre
générée et la base SQLite. Conservez-le : perdre la clé maîtresse rend le coffre illisible.

## Sécurité des comptes locaux

- Hachages de mots de passe Argon2id, sels aléatoires et mot de passe d'au moins 12 octets.
- Sessions navigateur aléatoires, stockées sous forme de hachages, expirant après 12 heures ; cookies HttpOnly/SameSite
  et cookies Secure en HTTPS. La déconnexion révoque la session navigateur côté serveur.
- Les changements de mot de passe, la désactivation de compte et les changements de rôle révoquent les sessions navigateur et les appareils appairés.
- Limitation des tentatives de connexion par compte et vérifications de mots de passe concurrentes bornées. Les connexions incorrectes et
  limitées renvoient la même erreur et effectuent le même travail Argon2.
- Le dernier administrateur actif ne peut être ni désactivé ni rétrogradé.
- La création de compte, les mises à jour de rôle et l'appairage mobile revérifient la session courante dans la mutation de base de données,
  y compris pour les requêtes déjà en cours lorsqu'un compte a été révoqué.

Tous les utilisateurs autorisés peuvent utiliser le coffre partagé et administrer les machines configurées. Les administrateurs
gèrent aussi les comptes et les sauvegardes de contenu de l'instance. Un appareil mobile appairé ne peut ni gérer les comptes,
ni appairer d'autres appareils, ni exporter/restaurer des sauvegardes, même si son propriétaire est administrateur.
Il n'y a pas d'inscription publique. Gardez au moins deux comptes administrateurs pour le dépannage.

## SSO optionnel

Utilisez `compose.sso.yaml` avec Compose **2.24.4 ou plus récent**. Copiez `.env.sso.example` vers `.env`, renseignez
l'émetteur OIDC, l'ID client et le secret client, des `RELAY_ALLOWED_USERS` et `RELAY_ADMIN_USERS` explicites,
et générez `OAUTH2_PROXY_COOKIE_SECRET` avec `openssl rand -base64 32`. Protégez `.env` avec
`chmod 600 .env`. Enregistrez `${RELAY_PUBLIC_ORIGIN}/oauth2/callback` auprès du fournisseur.

```sh
docker compose -f compose.yaml -f compose.sso.yaml pull
docker compose -f compose.yaml -f compose.sso.yaml up -d
```

Entra ID, Keycloak ou Authentik peuvent fournir l'OIDC. Un proxy HTTPS doit rediriger vers
`127.0.0.1:4180`. Dans ce mode, Relay valide la session SSO via oauth2-proxy ; l'écran de connexion local
et la gestion des comptes locaux sont désactivés. Le SSO et l'authentification locale sont deux modes
d'installation alternatifs. Changer de mode ne supprime ni les comptes ni l'inventaire existants.

Pour un oauth2-proxy existant, définissez `RELAY_AUTH_MODE=userinfo` et `RELAY_OAUTH2_USERINFO_URL`
et connectez Relay à son réseau. L'authentification par en-têtes de confiance exige aussi des CIDR
`RELAY_TRUSTED_PROXIES` explicites. Conservez les routes publiques des métadonnées PWA dans
`deploy/nginx-relay-pwa.conf`, et laissez passer `/api/` pour l'authentification mobile par jeton (bearer).
Relay authentifie lui-même chaque requête d'API.

## Fonctionnalités

- Consoles SSH, RDP, VNC et Proxmox dans des onglets, plus les connexions rapides.
- Identifiants réutilisables chiffrés en AES-256-GCM et héritage par protocole le long des dossiers.
- SFTP, lecteur RDP redirigé par utilisateur, presse-papiers et plein écran.
- Synchronisation NetBox/Proxmox, aperçu d'import RDM et fusion des doublons.
- Commandes SSH parallèles avec sortie bornée, délai d'expiration et annulation.
- Notes Markdown, pièces jointes, espaces de travail privés et enregistrements.
- Contrôles de disponibilité TCP, filtres et nouveaux tests ciblés.
- Sauvegardes de contenu chiffrées par mot de passe, enregistrements terminés inclus.
- Comptes locaux ou SSO, appairage mobile révocable et PWA installable.

## Confiance et protection des données

SSH utilise la confiance à la première utilisation (TOFU) et épingle les clés des serveurs pour les sessions interactives et les commandes.
Vérifiez le premier contact sur un réseau de confiance. Une clé modifiée bloque les connexions suivantes.
L'exécuteur de commandes et guacd utilisent la même préférence d'algorithme lorsqu'un hôte propose plusieurs clés.
Après l'import de données plus anciennes, une empreinte existante utilisant un autre algorithme peut nécessiter une vérification
manuelle de l'empreinte avant la réinitialisation de cette empreinte. Les nouvelles entrées RDP valident les certificats par
défaut ; les dérogations explicites existantes sont conservées.

Les notes, pièces jointes, sorties de commandes et enregistrements sont des fichiers/données ordinaires dans le volume protégé.
Utilisez le chiffrement de disque si nécessaire. Le chiffrement du coffre ne protège pas contre la compromission de l'hôte
qui détient sa clé maîtresse. Les administrateurs de sauvegarde peuvent exporter les captures et espaces de travail privés de chaque utilisateur.

Les sauvegardes de contenu sont chiffrées séparément par mot de passe. La restauration remplace l'inventaire, le coffre et les
enregistrements terminés, et exige de fermer les sessions et d'attendre la fin des commandes/synchronisations. La limite de l'archive est de
512 Mio non compressés. **Les comptes locaux, les sessions navigateur et les appairages mobiles relèvent de la politique
d'accès de l'installation : la restauration de contenu les préserve.** Les jetons d'appareil des archives anciennes sont ignorés.
Utilisez une sauvegarde cohérente et protégée du volume de l'hôte pour préserver la politique des comptes en cas de reprise après sinistre.

Les répertoires de transfert RDP utilisent un espace de noms d'identité SHA-256. Les fichiers anciens doivent être déplacés par un
opérateur après avoir vérifié que l'ancien nom de répertoire n'était pas partagé par plusieurs identités.
N'exposez jamais guacd ni l'authentification de développement à un réseau non fiable.

## Images et mises à jour

Relay s'exécute en utilisateur non root dans une image distroless. L'image guacd fournie compile Apache Guacamole
1.6.0 à partir de sources vérifiées par somme de contrôle, sur une base Alpine maintenue, avec les bibliothèques système FreeRDP 3
et un rétroportage de compatibilité Apache PostConnect inclus (GUACAMOLE-2328).
La base obsolète de l'ancienne image guacd officielle n'est pas utilisée par l'installation recommandée.
Les deux images ciblent amd64 et arm64 et incluent une SBOM, la provenance de compilation et les mentions de licence.

Épinglez `RELAY_IMAGE` sur une version publiée pour des mises à jour reproductibles. Exportez une sauvegarde de contenu et conservez
une sauvegarde protégée du volume de l'hôte, puis exécutez `docker compose pull && docker compose up -d`. Conservez
les condensats (digests) des images précédentes et les sauvegardes pour un retour arrière. Extrayez les mises à jour dans le même répertoire
d'installation (ou conservez le même nom de projet Compose) afin que les volumes nommés existants soient réutilisés.
Ne supprimez jamais les volumes nommés lors des mises à jour.

## Compilation et développement

Nécessite Go 1.27.4 ou plus récent et Node.js 22 :

```sh
cd web && npm ci && npm run build && cd ..
go test ./...
docker compose -f compose.yaml -f compose.build.yaml build
```

L'authentification de développement doit être liée à la boucle locale :

```sh
RELAY_AUTH_MODE=dev RELAY_DEV_USER=developer@example.test RELAY_LISTEN=127.0.0.1:8080 go run ./cmd/relay
```

Voir `SECURITY_REVIEW.md` pour la couverture et les limites de la revue, et `SECURITY.md` pour les signalements privés.
Licence Apache-2.0 ; voir `LICENSE`, `NOTICE` et `THIRD_PARTY_NOTICES.md`.
Relay est indépendant de Devolutions et de Remote Desktop Manager.
