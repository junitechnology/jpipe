#!/bin/sh
find diagrams -type f -exec sh -c 'swirly -f {} assets/images/$(dirname {})/$(basename {} .txt).svg' \;