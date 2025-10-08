#!/bin/sh
# Copyright 2025 Richard Kosegi
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

YP_BIN="${YP_BIN:-yp}"

pipeline="pipeline.yaml"

if [[ "${1:-}" == "--file" ]]; then
  shift
  pipeline="${1}"
  shift
fi

vars=""
i=0
for file in "$@"; do
  vars="${vars} --set .vars.input"'['$i']'"=${file}"
  i=`expr $i + 1`
done

${YP_BIN} --file "${pipeline}" ${vars}
