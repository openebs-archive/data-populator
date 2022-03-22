#!/bin/sh
set -e

# Check and run if any script is available at /entrypoint.d path.
for f in /entrypoint.d/*; do
  case "$f" in
    *.sh)  echo "$0: running $f"; . "$f" ;;
    *)     echo "$0: ignoring $f" ;;
  esac
done
exec "$@"
