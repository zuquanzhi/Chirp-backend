#!/bin/bash

BASE_URL="http://localhost:9527"
EMAIL="admin_test_$(date +%s)@example.com"
PASSWORD="password123"
DB_DRIVER=${DB_DRIVER:-mysql}
DB_USER=${DB_USER:-chirp}
DB_PASS=${DB_PASS:-test12345}
DB_HOST=${DB_HOST:-127.0.0.1}
DB_NAME=${DB_NAME:-chirp}

if [ "$DB_DRIVER" != "mysql" ]; then
    echo "DB_DRIVER=$DB_DRIVER (non-MySQL). Admin promotion step will be skipped."
    SKIP_DB=1
fi

echo "1. Registering user..."
REG_RESP=$(curl -s -X POST "$BASE_URL/signup" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"Admin Tester\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
echo "Response: $REG_RESP"
USER_ID=$(echo $REG_RESP | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
echo "User ID: $USER_ID"

echo -e "\n\n2. Logging in..."
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
echo "Response: $LOGIN_RESP"

TOKEN=$(echo $LOGIN_RESP | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")
echo "Token: $TOKEN"

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
    echo "Failed to get token"
    exit 1
fi

echo -e "\n3. Uploading resource..."
# Create a dummy file
echo "This is a test file content for admin test $(date)" > test_resource.txt

UPLOAD_RESP=$(curl -s -X POST "$BASE_URL/api/public/resources" \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@test_resource.txt" \
    -F "title=Admin Test Resource" \
    -F "description=Testing admin review" \
    -F "subject=Testing" \
    -F "type=Test")
echo "Response: $UPLOAD_RESP"

RES_ID=$(echo $UPLOAD_RESP | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
FILE_HASH=$(echo $UPLOAD_RESP | python3 -c "import sys, json; print(json.load(sys.stdin)['file_hash'])")

echo "Resource ID: $RES_ID"
echo "File Hash: $FILE_HASH"

if [ -z "$RES_ID" ] || [ "$RES_ID" == "null" ]; then
    echo "Failed to upload resource"
    exit 1
fi

echo -e "\n4. Reviewing resource (Approving) - Expecting 403 Forbidden (User Role)..."
REVIEW_RESP=$(curl -s -w "%{http_code}" -o /dev/null -X POST "$BASE_URL/api/admin/resources/$RES_ID/review" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"status\":\"APPROVED\"}")

if [ "$REVIEW_RESP" == "403" ]; then
    echo "Access denied as expected (403 Forbidden)"
else
    echo "Unexpected response: $REVIEW_RESP (Expected 403)"
    exit 1
fi

if [ -z "$SKIP_DB" ]; then
    echo -e "\n5. Promoting user to ADMIN..."
    mysql -u"$DB_USER" -p"$DB_PASS" -h"$DB_HOST" "$DB_NAME" -e "UPDATE users SET role='ADMIN' WHERE id=$USER_ID;"
    echo "User promoted."
else
    echo "Skipped DB promotion (non-MySQL)."
fi

if [ -z "$SKIP_DB" ]; then
    echo -e "\n6. Reviewing resource (Approving) - Expecting 200 OK (Admin Role)..."
    REVIEW_RESP=$(curl -s -w "%{http_code}" -o /dev/null -X POST "$BASE_URL/api/admin/resources/$RES_ID/review" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"status\":\"APPROVED\"}")

    if [ "$REVIEW_RESP" == "200" ]; then
        echo "Review successful (200 OK)"
    else
        echo "Review failed with status $REVIEW_RESP"
        exit 1
    fi
else
    echo "Skipped admin review (non-MySQL)."
fi

if [ -z "$SKIP_DB" ]; then
    echo -e "\n7. Checking duplicates..."
    DUP_RESP=$(curl -s -X GET "$BASE_URL/api/admin/resources/duplicates?hash=$FILE_HASH" \
        -H "Authorization: Bearer $TOKEN")
    echo "Response: $DUP_RESP"
fi

# Cleanup
rm test_resource.txt
