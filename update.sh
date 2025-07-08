#!/bin/bash -e

## Modules options

ORG=nrdcg

BASE_MAJOR_VERSION=65
BASE_OCI_VERSION=${BASE_MAJOR_VERSION}.95.1

OLD_MAJOR_VERSION=v${BASE_MAJOR_VERSION}

NEW_MAJOR_VERSION=v10${BASE_MAJOR_VERSION}
TAG=v10${BASE_OCI_VERSION}

## Clone options

OCI_VERSION=v${BASE_OCI_VERSION}
# SOURCE="/tmp/oci-go-sdk-clone/"
SOURCE=$(mktemp -d)
# DEST="/tmp/oci-go-sdk-fork"
DEST=$(mktemp -d)
DEST_BRANCH="modules"

## Clone original repository

rm -rf ${SOURCE}
git clone -c advice.detachedHead=false -q --branch ${OCI_VERSION} --single-branch --depth 1 git@github.com:oracle/oci-go-sdk.git ${SOURCE}
cd ${SOURCE}

CUR_GO=$(go mod edit -json | jq -r .Go)

## Quick fix existing code

cat > .golangci.yml <<EOF
version: "2"

formatters:
    enable:
      - gci

linters:
    default: none
    enable:
      - govet
    settings:
        govet:
          disable-all: true
          enable:
            - printf
EOF

go mod edit -go 1.24.0 -toolchain=none
go mod tidy

command -v golangci-lint >/dev/null 2>&1 || { echo "It requires golangci-lint but it's not installed. Aborting." >&2; exit 1; }

golangci-lint fmt
golangci-lint run --fix

rm .golangci.yml
git checkout HEAD go.mod go.sum

############################################
##              Modules                   ##
############################################

## Move helpers from examples package

#go install github.com/vikstrous/mvpkg@latest
command -v mvpkg >/dev/null 2>&1 || { echo "It requires github.com/vikstrous/mvpkg but it's not installed. Aborting." >&2; exit 1; }

mvpkg ./example/helpers ./helpers

## Format existing code

# --enable gci
golangci-lint fmt

## Find package dependencies

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd' | grep -v 'common'); do
    cd ${row}

    echo "Dependencies of ${row}"

    deps=$(go list -test -f '{{join .Imports "\n"}}' ./... \
         | grep "github.com/oracle/oci-go-sdk/${OLD_MAJOR_VERSION}" \
         | sort -u \
         | sed "s|/${OLD_MAJOR_VERSION}/|/|" \
         | awk -F/ "BEGIN{OFS=\"/\"} {print \$1, \"${ORG}\", \$3, \$4, \"${NEW_MAJOR_VERSION}\"}" \
         | sort \
         | uniq)

    echo "${deps}" > "${row}.txt"

    cd ..
done

## Init modules

echo "create module github.com/${ORG}/oci-go-sdk/common/${NEW_MAJOR_VERSION}"

cd common

go mod init github.com/${ORG}/oci-go-sdk/common/${NEW_MAJOR_VERSION}
go mod edit -go ${CUR_GO} -toolchain=none
go mod tidy
go mod edit -toolchain=none

cd ..

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd' | grep -v 'common'); do
    echo "create module github.com/${ORG}/oci-go-sdk/${row}/${NEW_MAJOR_VERSION}"

    cd ${row}

    go mod init github.com/${ORG}/oci-go-sdk/${row}/${NEW_MAJOR_VERSION}
    go mod edit -go ${CUR_GO} -toolchain=none

    for dep in $(cat "${row}.txt"); do
        target=$(echo ${dep} | awk -F/ 'BEGIN{OFS="/"} {print $4}')
        if [ -z "${target}" ]; then
          continue
        fi

        echo "replace ${dep}: target=${target} row=${row}"

        if [[ "${target}" != "${row}" ]]; then
            go mod edit -replace ${dep}=$(echo ${dep} | awk -F/ 'BEGIN{OFS="/"} {print "..", $4}')
        else
            echo "SKIP: ${dep} for ${row}"
        fi
    done

    rm "${row}.txt"

    cd ..
done

## Replace package with module

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd'); do
    echo "Replace package with module ${row}"

    find . -type f -name "*.go" \
      -exec grep -l "github.com/oracle/oci-go-sdk/${OLD_MAJOR_VERSION}/${row}" {} + \
      | xargs -r -d '\n' -I{} sed -i \
        -e "s|github.com/oracle/oci-go-sdk/${OLD_MAJOR_VERSION}/${row}\"|github.com/${ORG}/oci-go-sdk/${row}/${NEW_MAJOR_VERSION}\"|g" \
        -e "s|github.com/oracle/oci-go-sdk/${OLD_MAJOR_VERSION}/${row}/|github.com/${ORG}/oci-go-sdk/${row}/${NEW_MAJOR_VERSION}/|g" {}

done

## go tidy

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd'); do
    echo "Tidy ${row}"

    cd ${row}

    go mod tidy
    go mod edit -toolchain=none

    cd ..
done

## Replace version inside go.mod files

find . -type f -name "go.mod" \
    -exec grep -l "github.com/${ORG}/oci-go-sdk/" {} + \
    | xargs -r -d '\n' -I{} sed -i \
    -E "s|(github.com/${ORG}/oci-go-sdk/.+) v(.+)|\1 ${TAG}|g" {}

## Extra: remove version generator

if [[ "${ORG}" != "oracle" ]]; then
  rm -rf cmd
  rm oci.go go.mod go.sum
fi

## Remove git data.

rm -rf .git

############################################
##              Commit                    ##
############################################

## Prepare the fork
# git clone -q --branch master --single-branch git@github.com:nrdcg/oci-go-sdk.git /tmp/oci-go-sdk-clone
# cd /tmp/oci-go-sdk-clone
# git checkout --orphan ${DEST_BRANCH}
# git rm -rf .
# git commit -m "chore: initial empty commit." --allow-empty
# git push origin ${DEST_BRANCH}

rm -rf ${DEST}

git clone -q --branch ${DEST_BRANCH} --single-branch git@github.com:${ORG}/oci-go-sdk.git ${DEST}

cd ${DEST}

git rm -f -r --ignore-unmatch '*'

cp -r ${SOURCE}/. .

git add .

git commit -m "feat: modules oci-go-sdk ${OCI_VERSION}"

git push origin ${DEST_BRANCH}

# A new module version may be published by pushing a tag to the repository that contains the module source code.
# The tag is formed by concatenating two strings: a prefix and a version.
# https://go.dev/wiki/Modules#publishing-a-release

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd'| grep -v 'example'); do
    git tag ${row}/${TAG}
    git push origin ${row}/${TAG}
done

### Ignored for the fork
#git tag ${VERSION}
#git push origin ${VERSION}

find . -name go.mod -execdir go list -f "- \`{{.ImportPath}} v${NEW_MAJOR_VERSION}\`" \; | grep -v example | sort > modules.md

cd -

rm -rf ${SOURCE}
rm -rf ${DEST}
