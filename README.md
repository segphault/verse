# Verse

Verse is a lightweight rule-based reverse proxy written in Go. It makes it easy to setup subdomains that expose services running inside of Docker containers. It can also optionally serve static files. It has built-in SNI support so you can configure TLS individually for each matching rule.
