# syntax=docker/dockerfile:1

# ---------- Stage 1: build (Go, Node, protoc; builds the UI and the binary) ----------
FROM golang:1.27.2-trixie AS build

# Node.js and npm for "make ui" (only in this stage, not in the final image)
COPY --from=node:26.4.0-trixie-slim /usr/local/bin/node /usr/local/bin/node
COPY --from=node:26.4.0-trixie-slim /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
 && node --version && npm --version

# protoc: needed by the version check in "make ui"
# libcap2-bin: provides setcap, used below for ports 443 and 80
RUN apt-get update \
 && apt-get install -y --no-install-recommends protobuf-compiler libcap2-bin \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /src

# Go dependencies first (cached while go.mod and go.sum do not change)
COPY go.mod go.sum go.work ./
COPY kit/go.mod kit/go.sum ./kit/
RUN go mod download
# protoc-gen-go at the version pinned in go.mod (needed by the version check)
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go

# npm dependencies (cached while package-lock.json does not change)
COPY public/package.json public/package-lock.json ./public/
RUN cd public && npm ci

# Rest of the source code (.dockerignore keeps out databases, secrets, node_modules)
COPY . .

# Frontend: "make ui" needs a .env file to exist.
# cmd/esbuild ignores its own errors, so fail here if no bundle was produced.
RUN touch .env \
 && make ui \
 && ls public/dist/*.js > /dev/null \
 && test -f public/assets/index.html

# Server binary (CGO is needed by the SQLite driver), then allow the
# non-root user to open ports 443 and 80 (same as "make install" upstream)
RUN CGO_ENABLED=1 go build -trimpath -o /out/quickfeed . \
 && setcap cap_net_bind_service=+ep /out/quickfeed

# qf-dbcopy: one-time copy of the SQLite database into PostgreSQL
# (used by deploy/k8s/jobs/migrate-sqlite-to-postgres.yaml)
RUN CGO_ENABLED=1 go build -trimpath -o /out/qf-dbcopy ./cmd/dbcopy


# ---------- Stage 2: runtime (only what QuickFeed needs to run) ----------
FROM debian:trixie-20261005-slim AS runtime

# CA certificates: trust HTTPS connections to GitHub and Let's Encrypt
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# Non-root user. Only /app/logs and /data are writable.
# /app/.env stays empty and read-only: all configuration comes from the environment.
# /data is empty: the database is mounted here at run time, never baked in.
RUN groupadd --gid 10001 quickfeed \
 && useradd --uid 10001 --gid 10001 --home-dir /data \
            --no-create-home --shell /usr/sbin/nologin quickfeed \
 && mkdir -p /app/logs/courses /data \
 && touch /app/.env \
 && chown -R 10001:10001 /app/logs /data

ENV QUICKFEED=/app \
    HOME=/data

COPY --from=build /out/quickfeed     /usr/local/bin/quickfeed
COPY --from=build /out/qf-dbcopy     /usr/local/bin/qf-dbcopy
COPY --from=build /src/public/assets /app/public/assets
COPY --from=build /src/public/dist   /app/public/dist

USER 10001:10001
EXPOSE 443 80
ENTRYPOINT ["/usr/local/bin/quickfeed"]
CMD ["-database.file=/data/qf.db"]