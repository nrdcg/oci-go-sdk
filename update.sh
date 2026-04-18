#!/bin/bash -e

INITIAL_PATH=$(pwd)

## Modules options

SRC_ORG=oracle
DEST_ORG=nrdcg

REPO_NAME=oci-go-sdk

LIB_VERSION=$(curl -s https://api.github.com/repos/${SRC_ORG}/${REPO_NAME}/releases/latest | jq -r '.tag_name')

TRIMMED_VERSION="${LIB_VERSION#[vV]}"

# SRC_BASE_MAJOR_VERSION=65
SRC_BASE_MAJOR_VERSION="${TRIMMED_VERSION%%.*}"
# SRC_BASE_VERSION=${SRC_BASE_MAJOR_VERSION}.97.1
SRC_BASE_VERSION="${SRC_BASE_MAJOR_VERSION}.${TRIMMED_VERSION#*.}"

SRC_MAJOR_VERSION=v${SRC_BASE_MAJOR_VERSION}
DEST_MAJOR_VERSION=v10${SRC_BASE_MAJOR_VERSION}

SRC_TAG=v${SRC_BASE_VERSION}
DEST_TAG=v10${SRC_BASE_VERSION}

## Clone options

SRC_DIR=$(mktemp -d)
DEST=$(mktemp -d)

DEST_BRANCH="modules"

SRC_REMOTE="git@github.com:${SRC_ORG}/${REPO_NAME}.git"
DEST_REMOTE="git@github.com:${DEST_ORG}/${REPO_NAME}.git"

################

## Prepare the fork
# git clone -q --single-branch ${DEST_REMOTE} /tmp/${REPO_NAME}-clone
# cd /tmp/${REPO_NAME}-clone
# git checkout --orphan ${DEST_BRANCH}
# git rm -rf .
# git commit -m "chore: initial empty commit." --allow-empty
# git push origin ${DEST_BRANCH}
#
# exit 0

################

echo "Update to ${LIB_VERSION}"

## Clone original repository

rm -rf ${SRC_DIR}
git clone -c advice.detachedHead=false -q --branch ${SRC_TAG} --single-branch --depth 1 ${SRC_REMOTE} ${SRC_DIR}
cd ${SRC_DIR}

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
         | grep "github.com/${SRC_ORG}/${REPO_NAME}/${SRC_MAJOR_VERSION}" \
         | sort -u \
         | sed "s|/${SRC_MAJOR_VERSION}/|/|" \
         | awk -F/ "BEGIN{OFS=\"/\"} {print \$1, \"${DEST_ORG}\", \$3, \$4, \"${DEST_MAJOR_VERSION}\"}" \
         | sort \
         | uniq)

    echo "${deps}" > "${row}.txt"

    cd ..
done

## Init modules

echo "create module github.com/${DEST_ORG}/${REPO_NAME}/common/${DEST_MAJOR_VERSION}"

cd common

go mod init github.com/${DEST_ORG}/${REPO_NAME}/common/${DEST_MAJOR_VERSION}
go mod edit -go ${CUR_GO} -toolchain=none
go mod tidy
go mod edit -toolchain=none

cd ..

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd' | grep -v 'common'); do
    echo "create module github.com/${DEST_ORG}/${REPO_NAME}/${row}/${DEST_MAJOR_VERSION}"

    cd ${row}

    go mod init github.com/${DEST_ORG}/${REPO_NAME}/${row}/${DEST_MAJOR_VERSION}
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

    src_import="github.com/${SRC_ORG}/${REPO_NAME}/${SRC_MAJOR_VERSION}/${row}"
    dest_import="github.com/${DEST_ORG}/${REPO_NAME}/${row}/${DEST_MAJOR_VERSION}"

    find . -type f -name "*.go" \
      -exec grep -l "${src_import}" {} + \
      | xargs -r -d '\n' -I{} sed -i \
        -e "s|${src_import}\"|${dest_import}\"|g" \
        -e "s|${src_import}/|${dest_import}/|g" {}

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
    -exec grep -l "github.com/${DEST_ORG}/${REPO_NAME}/" {} + \
    | xargs -r -d '\n' -I{} sed -i \
    -E "s|(github.com/${DEST_ORG}/${REPO_NAME}/.+) v(.+)|\1 ${DEST_TAG}|g" {}

## Extra: remove version generator

if [[ "${DEST_ORG}" != "${SRC_ORG}" ]]; then
  rm -rf cmd
  rm oci.go go.mod go.sum
fi

## Copy LICENSE.txt into each submodule root (for license scanners / module zips)

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd'); do
  if [[ -f "${row}/go.mod" ]]; then
    cp LICENSE.txt "${row}/LICENSE.txt"
  fi
done

## Remove git data.

rm -rf .git

############################################
##              Commit                    ##
############################################

rm -rf ${DEST}

git clone -q --branch ${DEST_BRANCH} --single-branch ${DEST_REMOTE} ${DEST}

cd ${DEST}

git rm -f -r --ignore-unmatch '*'

cp -r ${SRC_DIR}/. .

git add .

git commit -q -m "feat: modules ${REPO_NAME} ${SRC_TAG}"

git push -q origin ${DEST_BRANCH}

# A new module version may be published by pushing a tag to the repository that contains the module source code.
# The tag is formed by concatenating two strings: a prefix and a version.
# https://go.dev/wiki/Modules#publishing-a-release

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd'| grep -v 'example'); do
    echo "tag: ${row}/${DEST_TAG}"
    git tag ${row}/${DEST_TAG}
    git push -q origin ${row}/${DEST_TAG}
done

### Ignored for the fork
#git tag ${VERSION}
#git push origin ${VERSION}

find . -name go.mod -execdir go list -f "- \`{{.ImportPath}} ${DEST_TAG}\`" \; | grep -v example | sort > ${INITIAL_PATH}/modules.md

cd -

rm -rf ${SRC_DIR}
rm -rf ${DEST}
