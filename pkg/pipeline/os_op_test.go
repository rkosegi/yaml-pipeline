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

package pipeline

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOsOp(t *testing.T) {
	t.Run("mkdir", func(t *testing.T) {
		t.Run("Invalid empty dir", func(t *testing.T) {
			oo := &OsOpSpec{Mkdir: &OsOpMkdirSpec{
				Mode: ptr(os.FileMode(0o775)),
				Path: ValOrRef{},
			}}
			t.Log(oo)
			assert.Error(t, oo.Do(mockEmptyActCtx()))
		})
		t.Run("Create valid in the loop", func(t *testing.T) {
			fe := ForEachOpSpec{
				Item: &ValOrRefSlice{
					&ValOrRef{Val: t.TempDir() + "/mkdir1"},
					&ValOrRef{Val: t.TempDir() + "/mkdir2"},
				},
				Action: ActionSpec{
					Operations: OpSpec{
						Os: &OsOpSpec{
							Mkdir: &OsOpMkdirSpec{
								Path: ValOrRef{
									Val: "{{ .forEach }}",
								},
							},
						},
					},
				},
			}
			t.Log(fe.String())
			assert.NoError(t, fe.Do(mockEmptyActCtx()))
		})
	})
	t.Run("chmod", func(t *testing.T) {
		t.Run("Invalid empty dir", func(t *testing.T) {
			oo := &OsOpSpec{Chmod: &OsOpChmodSpec{}}
			t.Log(oo)
			assert.Error(t, oo.Do(mockEmptyActCtx()))
		})
	})

	t.Run("chdir", func(t *testing.T) {
		t.Run("Invalid empty dir", func(t *testing.T) {
			oo := &OsOpSpec{Chdir: &OsOpChdirSpec{}}
			t.Log(oo)
			assert.Error(t, oo.Do(mockEmptyActCtx()))
		})
		t.Run("valid", func(t *testing.T) {
			d, err := os.Getwd()
			t.Cleanup(func() {
				_ = os.Chdir(d)
			})
			assert.NoError(t, err)
			oo := &OsOpSpec{Chdir: &OsOpChdirSpec{Path: ValOrRef{Val: t.TempDir()}}}
			t.Log(oo)
			assert.NoError(t, oo.Do(mockEmptyActCtx()))
		})
	})
}
