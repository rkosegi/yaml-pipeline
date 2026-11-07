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

Library          Process
Resource         libraries/yp.resource

Suite Setup      Setup


*** Test Cases ***
Print version
    [Documentation]     Check if CLI can print version
    Builtin.Log    Checking CLI version
    yp.Run CLI    --version

Attempt To Run Non Existing Pipeline
    [Documentation]     Try to run pipeline from a file that don't exists.
    Builtin.Run Keyword And Expect Error    2 != 0
    ...    yp.Run CLI
    ...    --file /this/does/not/exists

Run Simplest Pipeline Possible
    [Documentation]     Run a pipeline that just prints a message
    yp.Run CLI    --file testdata/pipeline_just_msg.yaml

Run Pipeline That Just Fail
    [Documentation]     Run a pipeline that purposely fail
    Builtin.Run Keyword And Expect Error    2 != 0
    ...    yp.Run CLI
    ...    --file testdata/pipeline_just_fail.yaml


*** Keywords ***
Setup
    [Documentation]         Setup this suite
    Builtin.Log    Setting up suite
    Process.Run Process    make
