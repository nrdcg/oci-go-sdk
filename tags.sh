#!/bin/bash -e

TAG=v1065.95.0

# A new module version may be published by pushing a tag to the repository that contains the module source code.
# The tag is formed by concatenating two strings: a prefix and a version.
# https://go.dev/wiki/Modules#publishing-a-release

for row in $(ls -d */ | sed 's|[/]||g' | grep -v 'cmd'| grep -v 'example'); do
    git tag ${row}/${TAG}
    git push origin ${row}/${TAG}
done

### Ignored for the fork
git tag ${VERSION}
git push origin ${VERSION}
