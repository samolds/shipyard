#!/bin/sh
# sed_env_vars.sh
#
# run before starting the nginx server in this docker image to allow for a
# configurable api server host at runtime

set -e

SED_DIR="$1"

if [[ ! -z "${SED_DIR}" ]]; then

  DEV_PUBLIC_API_URL="http://localhost:8080"
  if [[ ! -z "${PUBLIC_API_URL}" ]]; then
    find ${SED_DIR} -type f -exec \
      sed -i -e 's@'"${DEV_PUBLIC_API_URL}"'@'"${PUBLIC_API_URL}"'@g' {} \;
    echo "Backend address updated from ${DEV_PUBLIC_API_URL} to ${PUBLIC_API_URL}"
  fi

fi
