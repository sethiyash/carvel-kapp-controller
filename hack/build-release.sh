#!/bin/bash

set -e -x -u

mkdir -p tmp/
mkdir -p .imgpkg/

# makes the get_kappctrl_ver function available (scrapes version from git tag)
source $(dirname "$0")/version-util.sh

export version="$(get_kappctrl_ver)"

sed -ri 's/^(\s*)(kapp-controller.carvel.dev/version\s*:\s*v0.0.0\s*$)/\1kapp-controller.carvel.dev/version: $version/' config/deployment.yml

cat config/deployment.yml

# sed -ri'.bak' -e  's/^(\s*)("kapp-controller.carvel.dev/version"\s*:\s*"v0.0.0"\s*$)/\1"kapp-controller.carvel.dev/version": env(version)/' config/deployment.yml

# sed -ri 's/^(\s*)(image\s*:\s*nginx\s*$)/\1image: apache/' file

ytt -f config/ -f config-release -v dev.kapp_controller_version="$(get_kappctrl_ver)" --data-values-env=KCTRL | kbld --imgpkg-lock-output .imgpkg/images.yml -f- > ./tmp/release.yml

# yq 'select(.kind == "Deployment") | .metadata.annotations."kapp-controller.carvel.dev/version" = env(version)' ./tmp/release.yml

shasum -a 256 ./tmp/release.yml

cat ./tmp/release.yml

echo SUCCESS
