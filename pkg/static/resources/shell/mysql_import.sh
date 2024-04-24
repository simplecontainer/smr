#!/bin/bash

DATABASE=$1

echo "ghostmgr: Importing sql file from the previous database."
export MYSQL_PWD=$(echo $MYSQL_ROOT_PASSWORD)
mysql -u root $DATABASE < /tmp/ghost.sql