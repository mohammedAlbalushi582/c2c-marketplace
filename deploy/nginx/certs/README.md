# TLS certificates

Place your existing certificate files here (this directory is mounted read-only
into the nginx container at `/etc/nginx/certs`):

```
deploy/nginx/certs/fullchain.pem   # certificate + intermediate chain
deploy/nginx/certs/privkey.pem     # private key
```

If your files have different names, either rename them or update the
`ssl_certificate` / `ssl_certificate_key` paths in `deploy/nginx/conf.d/app.conf`.

The `.pem` / `.key` files are gitignored so your private key is never committed.
