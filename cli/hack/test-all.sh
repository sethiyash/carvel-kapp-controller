#!/bin/bash

set -e -x -u

./hack/build.sh

export KCTRL_BINARY_PATH="$PWD/kctrl"

# Enable to debug tests using prompt output in the workflow
# export KCTRL_DEBUG_BUFERED_OUTPUT_TESTS=true

# ./hack/test.sh
cd test/e2e
go test -run TestE2EInitAndReleaseCases
# ./hack/test-e2e.sh

echo ALL SUCCESS
