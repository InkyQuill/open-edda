#!/bin/sh
set -eu
# Run as root on yggdrasil after staging these files and image.tar in this folder.
[ "$(id -u)" = 0 ] || { echo 'Run with sudo on yggdrasil.' >&2; exit 1; }
base=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$base"
sha256sum -c SHA256SUMS
nginx -t
/usr/local/bin/k3s ctr images import --platform linux/amd64 "$base/image.tar"
available=/etc/nginx/sites-available/edda.inky.su
enabled=/etc/nginx/sites-enabled/edda.inky.su
if [ -e "$available" ] || [ -L "$available" ] || [ -e "$enabled" ] || [ -L "$enabled" ]; then
    echo 'Edda nginx configuration already exists; refusing to overwrite it.' >&2
    exit 1
fi
install -m 644 "$base/edda.inky.su.nginx.conf" "$available"
ln -s "$available" "$enabled"
if ! nginx -t; then
    rm "$enabled" "$available"
    echo 'New Edda configuration removed; existing nginx was not reloaded.' >&2
    exit 1
fi
systemctl reload nginx
printf 'Edda image imported; HTTPS host enabled.\n'
