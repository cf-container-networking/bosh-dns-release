#!/usr/bin/env bash

set -exu


export GOPATH=$PWD/dns-release
export PATH="${GOPATH}/bin":$PATH

go install bosh-dns/vendor/github.com/onsi/ginkgo/ginkgo

pushd $GOPATH/src/bosh-dns/performance_tests
    ginkgo -r -randomizeAllSpecs -randomizeSuites -race .
popd
