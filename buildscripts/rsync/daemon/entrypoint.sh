#!/bin/sh
# Copyright © 2022 The OpenEBS Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# Use the environment variables for rsync username and password.
# If not provided, the default values will be used
RSYNC_USERNAME=${RSYNC_USERNAME:-openebs-user}
RSYNC_PASSWORD=${RSYNC_PASSWORD:-openebs-pass}
echo "$RSYNC_USERNAME:$RSYNC_PASSWORD" >  /etc/rsyncd.secrets
chmod 600 /etc/rsyncd.secrets

# Check and run if any script is available at /entrypoint.d path.
for f in /entrypoint.d/*; do
  # shellcheck disable=SC1090
  case "$f" in
    *.sh)  echo "$0: running $f"; . "$f" ;;
    *)     echo "$0: ignoring $f" ;;
  esac
done
exec "$@"
