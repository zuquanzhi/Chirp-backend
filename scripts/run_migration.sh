#!/bin/bash

DB_DRIVER=${DB_DRIVER:-mysql}
DB_USER=${DB_USER:-chirp}
DB_PASS=${DB_PASS:-test12345}
DB_HOST=${DB_HOST:-127.0.0.1}
DB_NAME=${DB_NAME:-chirp}

if [ "$DB_DRIVER" != "mysql" ]; then
    echo "DB_DRIVER=$DB_DRIVER not supported by this migration script (MySQL only)."
    exit 0
fi

echo "Running migration: scripts/migrate_v2_add_role.sql"
mysql -u"$DB_USER" -p"$DB_PASS" -h"$DB_HOST" "$DB_NAME" < scripts/migrate_v2_add_role.sql

if [ $? -eq 0 ]; then
    echo "Migration successful."
else
    echo "Migration failed."
fi
