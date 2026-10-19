#!/bin/sh
# Builds parse-dmarc from this checkout, serves it behind nginx under /dmarc/
# (default.conf) and loads it in headless Chrome (check.mjs). Exits non-zero
# unless the dashboard renders and every request it makes stays under /dmarc/.
set -eu
cd "$(dirname "$0")"

dc() { docker compose -p parse-dmarc-nginx-e2e --profile e2e "$@"; }
trap 'dc down -v --remove-orphans >/dev/null 2>&1' EXIT

dc up -d --build parse-dmarc nginx

# The app image is FROM scratch, with no shell to probe from, so wait for it
# from the nginx side, through the proxy.
i=0
until dc exec -T nginx wget -q -O /dev/null http://127.0.0.1/dmarc/api/statistics; do
  i=$((i + 1))
  if [ "$i" -ge 30 ]; then
    echo "parse-dmarc did not answer behind nginx after 30s" >&2
    dc logs >&2
    exit 1
  fi
  sleep 1
done

dc run --rm browser
