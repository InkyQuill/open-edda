# Build with frontend/dist as context and the deployed image as EDDA_BASE_IMAGE.
# This updates the web assets while retaining the server, CLI and migrations.
ARG EDDA_BASE_IMAGE=open-edda:20261006-files-sync
FROM ${EDDA_BASE_IMAGE}
COPY --chown=10001:10001 . /app/frontend/
