#!/bin/sh
set -e

# Use the environment variables for rsync username and password.
# If not provided, the default values will be used
RSYNC_USERNAME=${RSYNC_USERNAME:-user}
RSYNC_PASSWORD=${RSYNC_PASSWORD:-pass}
echo "$RSYNC_USERNAME:$RSYNC_PASSWORD" >  /etc/rsyncd.secrets
chmod 600 /etc/rsyncd.secrets

# Check and run if any script is available at /entrypoint.d path.
for f in /entrypoint.d/*; do
  case "$f" in
    *.sh)  echo "$0: running $f"; . "$f" ;;
    *)     echo "$0: ignoring $f" ;;
  esac
done
exec "$@"
