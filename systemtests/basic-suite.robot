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
Documentation    Basic test suite

Library          OperatingSystem
Library          Process
Library          String
Resource         libraries/yp.resource

Suite Setup      Setup


*** Test Cases ***
Print version
    [Documentation]     Check if CLI can print version
    Builtin.Log    Checking CLI version
    yp.Run CLI    --version

Attempt To Run Non Existing Pipeline
    [Documentation]     Try to run pipeline from a file that don't exists.
    yp.Run CLI    --file /this/does/not/exists    2

Run Simplest Pipeline Possible
    [Documentation]     Run a pipeline that just prints a message
    yp.Run CLI    --file testdata/pipeline_just_msg.yaml

Run Pipeline That Just Fail
    [Documentation]     Run a pipeline that purposely fail
    yp.Run CLI    --file testdata/pipeline_just_fail.yaml    1

Verify failure event creates exports
    [Documentation]     Run a pipeline that purposely fail and verify that lifecycle events are triggered
    ${rand}    String.Generate Random String    8
    ${dir}    Builtin.Catenate    SEPARATOR=/    ${TEMPDIR}    ${rand}
    OperatingSystem.Create Directory    ${dir}
    yp.Run CLI    --file testdata/pipeline-lifecycle-test.yaml --set vars.fail="true" --set vars.debugDir="${dir}"    1
    ${fileCount}    OperatingSystem.Count Files In Directory    ${dir}    debug-success.yaml
    Builtin.Should Be Equal    0    ${fileCount}    types=int
    ${fileCount}    OperatingSystem.Count Files In Directory    ${dir}    debug-complete.yaml
    Builtin.Should Be Equal    1    ${fileCount}    types=int
    ${fileCount}    OperatingSystem.Count Files In Directory    ${dir}    debug-failure.yaml
    Builtin.Should Be Equal    1    ${fileCount}    types=int
    [Teardown]    OperatingSystem.Remove Directory    ${dir}    ${True}

Verify success event creates exports
    [Documentation]     Run a pipeline that purposely fail and verify that lifecycle events are triggered
    ${rand}    String.Generate Random String    8
    ${dir}    Builtin.Catenate    SEPARATOR=/    ${TEMPDIR}    ${rand}
    OperatingSystem.Create Directory    ${dir}
    yp.Run CLI
    ...    --file testdata/pipeline-lifecycle-test.yaml --set vars.fail="false" --set vars.debugDir="${dir}"
    ...    0
    ${fileCount}    OperatingSystem.Count Files In Directory    ${dir}    debug-success.yaml
    Builtin.Should Be Equal    1    ${fileCount}    types=int
    ${fileCount}    OperatingSystem.Count Files In Directory    ${dir}    debug-complete.yaml
    Builtin.Should Be Equal    1    ${fileCount}    types=int
    ${fileCount}    OperatingSystem.Count Files In Directory    ${dir}    debug-failure.yaml
    Builtin.Should Be Equal    0    ${fileCount}    types=int
    [Teardown]    OperatingSystem.Remove Directory    ${dir}    ${True}


*** Keywords ***
Setup
    [Documentation]         Setup this suite
    Builtin.Log    Setting up suite
    Builtin.Log Variables
    Process.Run Process    make
