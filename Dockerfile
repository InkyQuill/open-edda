FROM oven/bun:1.4.0@sha256:5ff609364c049b54eb0ff560ec96319729a972078ef2c755d758f0c6ef89c2d6 AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/bun.lock frontend/.npmrc ./
RUN bun install --frozen-lockfile
COPY frontend/ ./
RUN bun run build

FROM golang:1.27.1-bookworm@sha256:a4f46dc39c6b0359a3e1ed86ef14d01b374cc808649679dd5fca2290e6d54202 AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -tags sqlite_fts5 -trimpath -o /out/open-edda . && \
    CGO_ENABLED=1 go build -tags sqlite_fts5 -trimpath -o /out/edda ./cmd/edda

FROM debian:bookworm-slim@sha256:7c7b2c966bc9ee8cedfeef67e0e279108992c77681fa595db4a9d65c06ccc587
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    groupadd --gid 10001 edda && useradd --uid 10001 --gid edda --no-create-home edda && \
    mkdir -p /data /app && chown edda:edda /data
WORKDIR /app
COPY --from=backend /out/open-edda /out/edda /usr/local/bin/
COPY --from=frontend /src/frontend/dist /app/frontend
COPY migrations /app/migrations
ENV OPEN_EDDA_ADDR=0.0.0.0:8080 OPEN_EDDA_DATA_DIR=/data OPEN_EDDA_DB_PATH=/data/edda.db \
    OPEN_EDDA_MIGRATIONS_PATH=/app/migrations OPEN_EDDA_STATIC_PATH=/app/frontend
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["open-edda"]
