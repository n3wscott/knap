#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Only deepcopy the Duck types, as they are not real resources.
deepcopy-gen github.com/n3wscott/knap/pkg/apis/duck:v1alpha1
