#!/bin/bash

PRODUCT="dhcli"

set -euo pipefail

# Don't echo passwords.
set +x

cd "$(dirname "$0")" || exit 2


if [ "$ARTIFACT_VERSION" == "" ]; then
    echo "missing environment variable: ARTIFACT_VERSION" 1>&2
    exit 1
fi

BASE_URL="https://artifacts/artifactory/artifacts-dev-local/$PRODUCT/$ARTIFACT_VERSION"

if [ "$ARTIFACTORY_USERNAME" == "" ]; then
    echo "missing environment variable: ARTIFACTORY_USERNAME" 1>&2
    exit 1
fi
if [ "$ARTIFACTORY_PASSWORD" == "" ]; then
    echo "missing environment variable: ARTIFACTORY_PASSWORD" 1>&2
    exit 1
fi

echo "Uploading artifacts to $BASE_URL..."

cd artifacts/ || exit 2

#curl --user "$ARTIFACTORY_USERNAME:$ARTIFACTORY_PASSWORD" \
#    -X DELETE "${BASE_URL}/" || true

FILES="$(find . -type f | sed 's/^\.\///')"
for file in $FILES; do
    echo "* $file"
    curl --user "$ARTIFACTORY_USERNAME:$ARTIFACTORY_PASSWORD" \
        -T "$file" \
        "${BASE_URL}/${file}"
done
