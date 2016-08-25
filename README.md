# Verse

Verse is a lightweight rule-based reverse proxy written in Go. It makes it easy
to setup subdomains that expose services running inside of Docker containers.
It can also optionally serve static files. It has built-in support for fetching
certifcates from Let's Encrypt, which it will use on ports where TLS is
enabled.

## Using Docker

You can deploy Verse with Docker.
For this example, we use docker-compose.yaml, a Dockerfile, and a routes.json.
In docker-compose.yaml, we set up a pair of services, "fe" for "frontend" and
"nx" for "nginx".

```
version: '2'
services:
    fe:
        build: ./fe
        ports:
            - '80'
    nx:
        image: nginx
```

We use Verse as the frontend, and configure it with a routes.json.
The Dockerfile combines the verse Docker image (downloaded automatically from
Docker Hub) with the configuration.

```
FROM segphault/verse
COPY routes.json /routes.json
CMD verse /routes.json
EXPOSE 80
```

The configuration instructs Verse to listen on certain ports and
to forward traffic to other hosts if the incoming HTTP host header matches a
pattern.

```
{
    "servers": [
        {
            "port": 80,
            "rules": [
                {
                    "pattern": "example.com",
                    "binding": "nx"
                }
            ]
        }
    ]
}
```

The frontend can also serve static files from a given path within the Docker
container. Modify the Dockerfile to copy those assets.

```
FROM segphault/verse
COPY routes.json /routes.json
COPY www/ /www
CMD verse /routes.json
EXPOSE 80
```

Then, configure the route table with a nonempty "static" string.

```
{
    "servers": [
        {
            "port": 80,
            "static": "/www",
            "tls": false
        }
    ]
}
```

Verse can also terminate TLS on behalf of the proxied services.
The Dockerfile will need to expose the additional port.

```
FROM segphault/verse
COPY routes.json /routes.json
COPY www/ /www
CMD verse /routes.json
EXPOSE 80
EXPOSE 443
```

Marking a port with the "tls" flag enables encryption and the top-level "certs"
property points Verse at your certificates cache.

```
{
    "servers": [
        {
            "port": 443,
            "rules": [
                {
                    "pattern": "example.com",
                    "binding": "nx"
                }
            ],
            "static": "/www",
            "tls": true
        }
    ],
    "certs": "/letsencrypt.cache"
}
```

## Building and publishing to Docker

Cross-compile Verse for the Docker Linux environment.

```
GOOS=linux GOARCH=amd64 go build . -o verse-docker
```

Build the Docker image, tagged with the name of the published artifact.
The Docker image is based on Alpine linux and only retains the Verse binary.

```
docker build -t segphault/verse .
```

Publish the artifact to Docker Hub. Be sure to have an account and login from
the command line.

```
docker login
docker push segphault/verse
```
