#!/bin/sh

if [ $# -lt 3 ]; then
  echo "Usage: genjwt.sh <secret> <email> <level>"
fi

secret=$1
email=$2
level=${3:-l0}

payload_json='{"email":"'$email'","visibilityLevel":"'$level'"}'

#jwt_header=$(echo -n '{"alg":"HS256","typ":"JWT"}' | base64 | sed s/\+/-/g | sed 's/\//_/g' | sed -E s/=+$//)
jwt_header=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9

payload=$(echo -n $payload_json | base64 | sed s/\+/-/g |sed 's/\//_/g' |  sed -E s/=+$//)
hexsecret=$(echo -n "$secret" | xxd -p | paste -sd "")
hmac_signature=$(echo -n "${jwt_header}.${payload}" | openssl dgst -sha256 -mac HMAC -macopt hexkey:$hexsecret -binary | base64  | sed s/\+/-/g | sed 's/\//_/g' | sed -E s/=+$//)

echo "${jwt_header}.${payload}.${hmac_signature}"
