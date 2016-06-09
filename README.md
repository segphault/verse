# Verse

Verse is a lightweight rule-based reverse proxy written in Go. It makes it easy to setup subdomains that expose services running inside of Docker containers. It can also optionally serve static files. It has built-in support for fetching certifcates from Let's Encrypt, which it will use on ports where TLS is enabled.
