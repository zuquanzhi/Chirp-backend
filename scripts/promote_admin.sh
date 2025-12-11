#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <email>"
  echo "Example: $0 admin@example.com"
  exit 1
fi

EMAIL="$1"
DB_DRIVER=${DB_DRIVER:-mysql}
DB_USER=${DB_USER:-chirp}
DB_PASS=${DB_PASS:-test12345}
DB_HOST=${DB_HOST:-127.0.0.1}
DB_NAME=${DB_NAME:-chirp}

if [ "$DB_DRIVER" != "mysql" ]; then
    echo "This script only supports MySQL. Current DB_DRIVER=$DB_DRIVER"
    exit 1
fi

# Check if mysql client is available
if ! command -v mysql &> /dev/null; then
    echo "Error: 'mysql' command not found. Please install MySQL client."
    exit 1
fi

# Check if user exists
echo "Checking user..."
USER_COUNT=$(mysql -u"$DB_USER" -p"$DB_PASS" -h"$DB_HOST" "$DB_NAME" -N -s -e "SELECT count(*) FROM users WHERE email='$EMAIL'")

if [ "$USER_COUNT" -eq "0" ]; then
    echo "Error: User with email '$EMAIL' not found."
    exit 1
fi

# Update role
echo "Promoting user '$EMAIL' to ADMIN..."
mysql -u"$DB_USER" -p"$DB_PASS" -h"$DB_HOST" "$DB_NAME" -e "UPDATE users SET role='ADMIN' WHERE email='$EMAIL';"

if [ $? -eq 0 ]; then
    echo "Success! User '$EMAIL' is now an ADMIN."
else
    echo "Database operation failed."
fi
