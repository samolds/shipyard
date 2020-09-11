#!/bin/sh
# sed_env_vars.sh
#
# run before starting the nginx server in this docker image to allow for a
# configurable api server host at runtime

set -e

SED_DIR="$1"

if [[ ! -z "${SED_DIR}" ]]; then

  DEV_BACKEND_ADDRESS="http://localhost:8080"
  if [[ ! -z "${BACKEND_ADDRESS}" ]]; then
    find ${SED_DIR} -type f -exec \
      sed -i -e 's@'"${DEV_BACKEND_ADDRESS}"'@'"${BACKEND_ADDRESS}"'@g' {} \;
    echo "Backend address updated from ${DEV_BACKEND_ADDRESS} to ${BACKEND_ADDRESS}"
  fi

fi
