*** Comments ***
# Copyright 2026 Richard Kosegi
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


*** Settings ***
Documentation    YP keywords


*** Keywords ***
Run CLI
    [Documentation]     Run a CLI and expect exit code
    [Arguments]    ${args}    ${rc}=0
    ${result} =    Process.Run Process    ./dist/yp_linux_amd64_v1/yp --color never ${args}
    ...    shell=True
    Builtin.Log    ${result.stdout}
    Builtin.Log    ${result.stderr}
    Builtin.Log    Exit code: ${result.rc}
    Builtin.Should Be Equal    ${result.rc}    ${rc}    type=int
    RETURN    ${result}
