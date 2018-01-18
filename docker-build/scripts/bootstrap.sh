#!/bin/bash
# Main portal to start Ockle server.
#   Maintainer: xiaming.chen@transwarp.io

[ -f /external/scripts/init.sh ] && {
  . /external/scripts/init.sh
}

set -e

echo "options use-vc" >> /etc/resolv.conf

FILEPATH=$(cd ${0%/*} && echo $PWD/${0##*/})
THISFOLDER=$(cd $(dirname $FILEPATH) && pwd)

[ -f $THISFOLDER/env.sh ] && {
  . $THISFOLDER/env.sh
}

cd $WALM_HOME

WALM_CMD=`which walm_server`
if [ $? -ne 0 ]; then
    echo "Executable binary 'walm' not in the PATH, exit."
    exit 1
else
    echo "Found WALM binary $WALM_CMD"
fi

echo "Create database if not exists ..."
mysql -u ${WALM_MYSQL_USERNAME} \
    --password=${WALM_MYSQL_PASSWORD} \
    -h ${WALM_MYSQL_SERVER} \
    -P ${WALM_MYSQL_PORT} \
    -f << EOF
CREATE DATABASE IF NOT EXISTS ${WALM_DATABASE} DEFAULT CHARSET utf8 COLLATE utf8_general_ci;
EOF

# Initialize schema with migrated versions
$WALM_CMD db upgrade

$WALM_CMD runserver
