#!/bin/bash -e

ORG=oracle

OLD_MAJOR_VERSION=v65

NEW_MAJOR_VERSION=v1065
TAG=v1065.95.0

## Move helpers from examples package

#go install github.com/vikstrous/mvpkg@latest
mvpkg ./example/helpers ./helpers

## Format existing code

golangci-lint fmt --enable gci

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
go mod edit -go 1.13 -toolchain=none
go mod tidy
cd ..

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd' | grep -v 'common'); do
    echo "create module github.com/${ORG}/oci-go-sdk/${row}/${NEW_MAJOR_VERSION}"

    cd ${row}

    go mod init github.com/${ORG}/oci-go-sdk/${row}/${NEW_MAJOR_VERSION}
    go mod edit -go 1.13 -toolchain=none

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

## Replace

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd'); do
    echo "Replace ${row}"

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
  rm oci.go
  rm go.mod
  rm go.sum
fi

## Copy LICENSE.txt into each submodule root (for license scanners / module zips)

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd'); do
  if [[ -f "${row}/go.mod" ]]; then
    cp LICENSE.txt "${row}/LICENSE.txt"
  fi
done
