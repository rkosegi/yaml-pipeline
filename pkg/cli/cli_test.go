/*
Copyright 2025 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntrypoint(t *testing.T) {
	t.Run("no args should fail", func(t *testing.T) {
		cmd := New()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		assert.Error(t, cmd.Execute())
	})
	t.Run("--version should succeed", func(t *testing.T) {
		cmd := New()
		cmd.SetArgs([]string{"--version"})
		assert.NoError(t, cmd.Execute())
	})
	t.Run("invoking non-existent fail should fail", func(t *testing.T) {
		cmd := New()
		cmd.SetArgs([]string{"--file", "non-existent"})
		assert.Error(t, cmd.Execute())
	})
	t.Run("invoking known bad-behaving pipeline should fail", func(t *testing.T) {
		defer func() {
			recover()
		}()
		cmd := New()
		cmd.SetArgs([]string{"--file", "../../testdata/pipeline_just_fail.yaml"})
		assert.Error(t, cmd.Execute())
		t.Fail()
	})
	t.Run("invoking known OK pipeline should not fail", func(t *testing.T) {
		cmd := New()
		cmd.SetArgs([]string{"--file", "../../testdata/pipeline_just_msg.yaml"})
		assert.NoError(t, cmd.Execute())
	})

}
