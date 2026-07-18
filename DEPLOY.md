# Deployment guide (VPS + nginx + HTTPS)

Runs the full stack behind nginx, which terminates TLS with **your existing
certificate** and reverse-proxies to the app. Only ports **80/443** are public.

```
Internet ──443──► nginx ──► frontend (Next.js :3000)
                    ├──/api/────► backend (Go :8080)
                    └──/uploads/─► backend (Go :8080)
        postgres, strapi: internal only
```

## 0. Prerequisites (on the server)
- Docker + Docker Compose plugin
- A DNS **A record** pointing your domain at the server IP
- Firewall open for 80 + 443:
  ```bash
  ufw allow 80/tcp && ufw allow 443/tcp && ufw reload   # if you use ufw
  ```

## 1. Get the code
```bash
git clone git@github.com:mohammedAlbalushi582/c2c-marketplace.git
cd c2c-marketplace
```
(You've already done this.)

## 2. Install your TLS certificate
Copy your existing cert + key into `deploy/nginx/certs/` as:
```bash
deploy/nginx/certs/fullchain.pem   # cert + chain
deploy/nginx/certs/privkey.pem     # private key
```
(Different filenames? Either rename them or edit the `ssl_certificate*` paths in
`deploy/nginx/conf.d/app.conf`.) These files are gitignored.

## 3. Configure environment
```bash
cp .env.prod.example .env
nano .env        # set your domain in PUBLIC_BASE_URL and fill every CHANGE_ME
                 # tip: openssl rand -hex 32   for the JWT secrets
```

Generate the Strapi secrets file (`cms/.env`):
```bash
cat > cms/.env <<EOF
HOST=0.0.0.0
PORT=1337
APP_KEYS=$(openssl rand -base64 16),$(openssl rand -base64 16),$(openssl rand -base64 16),$(openssl rand -base64 16)
API_TOKEN_SALT=$(openssl rand -base64 16)
ADMIN_JWT_SECRET=$(openssl rand -base64 16)
TRANSFER_TOKEN_SALT=$(openssl rand -base64 16)
JWT_SECRET=$(openssl rand -base64 16)
ENCRYPTION_KEY=$(openssl rand -base64 16)
DATABASE_CLIENT=postgres
DATABASE_SSL=false
EOF
```
(The database host/name/user/password for Strapi come from the compose file, so
they don't need to be in `cms/.env`.)

## 4. (Recommended) Remove demo content for production
`migrations/000008_demo_content.*` seeds demo listings **and a demo admin
`admin@amjad.om` with a known password**. For a real site, delete it before the
first run so it never loads:
```bash
rm migrations/000008_demo_content.up.sql migrations/000008_demo_content.down.sql
```
The essential catalog (categories, custom fields, Oman locations) is in migration
`000007_seed` and stays.

## 5. Launch
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```
This builds all images, runs DB migrations automatically, and starts nginx.
Check status:
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
curl -k https://localhost/healthz         # {"db":"up","status":"ok"}
```
Open **https://your-domain.com** 🎉

## 6. Set up the Strapi admin
Strapi isn't public. Create its first admin via the CLI:
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec cms \
  npm run strapi -- admin:create-user --email you@example.com --password 'YourPass123' --firstname Admin
```
Then reach the panel through an SSH tunnel from your laptop:
```bash
ssh -L 1337:localhost:1337 root@your-server
# now open http://localhost:1337/admin
```

## 7. Updating later
```bash
git pull
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

## Handy alias
```bash
alias dcp='docker compose -f docker-compose.yml -f docker-compose.prod.yml'
dcp logs -f backend        # tail a service
dcp down                   # stop (keeps data); add -v to wipe the DB volume
```

## Notes & security
- Change/rotate the seeded marketplace admin password if you kept migration 8.
- Postgres and Strapi are **not** exposed on the public internet — only nginx is.
- For a public Strapi (subdomain like `cms.your-domain.com`), add a second server
  block to `app.conf` proxying to `cms:1337`, publish it, and point a cert at it.
- Uploaded images persist in the `uploads` Docker volume; the DB in `pgdata`.
  Back these up (`docker run --rm -v c2c-marketplace_pgdata:/data ...`).
