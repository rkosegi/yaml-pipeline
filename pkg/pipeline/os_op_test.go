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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rkosegi/yaml-toolkit/dom"
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
								Recursive: ptr(true),
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
	t.Run("userhome", func(t *testing.T) {
		x := os.Getenv("HOME")
		t.Run("invalid ", func(t *testing.T) {
			defer func() {
				_ = os.Setenv("HOME", x)
			}()
			_ = os.Unsetenv("HOME")
			oo := OsOpSpec{Userhome: &OsOpUserHomeSpec{StoreTo: ValOrRef{Val: "out"}}}
			t.Log(oo.String())
			assert.Error(t, oo.Do(mockEmptyActCtx()))
		})
		t.Run("valid ", func(t *testing.T) {
			gd := dom.ContainerNode()
			oo := OsOpSpec{Userhome: &OsOpUserHomeSpec{StoreTo: ValOrRef{Val: "out"}}}
			t.Log(oo.String())
			assert.NoError(t, oo.Do(newMockActBuilder().data(gd).build()))
			assert.Equal(t, x, gd.Child("out").AsLeaf().Value())

		})
	})
	t.Run("getcwd", func(t *testing.T) {
		t.Run("should get cwd", func(t *testing.T) {
			gd := dom.ContainerNode()
			oo := OsOpSpec{Getcwd: &OsOpGetcwdSpec{StoreTo: ValOrRef{Val: "out"}}}
			t.Log(oo.String())
			assert.NoError(t, oo.Do(newMockActBuilder().data(gd).build()))
			assert.True(t, len(gd.Child("out").AsLeaf().Value().(string)) > 0)

		})
	})
	t.Run("hostname", func(t *testing.T) {
		t.Run("should get hostname", func(t *testing.T) {
			gd := dom.ContainerNode()
			oo := OsOpSpec{Hostname: &OsOpHostnameSpec{StoreTo: ValOrRef{Val: "out"}}}
			t.Log(oo.String())
			assert.NoError(t, oo.Do(newMockActBuilder().data(gd).build()))
			assert.True(t, len(gd.Child("out").AsLeaf().Value().(string)) > 0)
		})
	})
	t.Run("readdir", func(t *testing.T) {
		t.Run("should get content of directory", func(t *testing.T) {
			x := t.TempDir()
			assert.NoError(t, os.WriteFile(filepath.Join(x, "file1.txt"), []byte("abcd"), 0777))
			gd := dom.ContainerNode()
			oo := OsOpSpec{
				Readdir: &OsOpReadDirSpec{
					Path: ValOrRef{
						Val: x,
					},
					StoreTo: ValOrRef{
						Val: "out",
					},
				}}
			t.Log(oo.String())
			assert.NoError(t, oo.Do(newMockActBuilder().data(gd).build()))
			assert.True(t, gd.Child("out").AsList().Size() > 0)

		})

		t.Run("read of invalid directory should fail", func(t *testing.T) {
			oo := OsOpSpec{
				Readdir: &OsOpReadDirSpec{
					Path: ValOrRef{
						Val: "/I/hope/this/does/not/exist",
					},
					StoreTo: ValOrRef{
						Val: "out",
					},
				}}
			t.Log(oo.String())
			assert.Error(t, oo.Do(mockEmptyActCtx()))
		})
	})
	t.Run("stat", func(t *testing.T) {
		t.Run("stat of temp directory should work", func(t *testing.T) {
			gd := dom.ContainerNode()
			oo := OsOpSpec{Stat: &OsOpStatSpec{
				Path: ValOrRef{Val: t.TempDir()}, StoreTo: ValOrRef{Val: "out"}},
			}
			t.Log(oo.String())
			assert.NoError(t, oo.Do(newMockActBuilder().data(gd).build()))
		})

		t.Run("stat of non-existent directory should fail", func(t *testing.T) {
			oo := OsOpSpec{Stat: &OsOpStatSpec{
				Path: ValOrRef{Val: filepath.Join(t.TempDir(), "12345678")}, StoreTo: ValOrRef{Val: "out"}},
			}
			t.Log(oo.String())
			assert.Error(t, oo.Do(mockEmptyActCtx()))
		})
	})
	t.Run("rename", func(t *testing.T) {
		t.Run("should rename file", func(t *testing.T) {
			f, err := os.CreateTemp("", "somefile")
			defer func(f *os.File) {
				_ = os.Remove(f.Name())
			}(f)
			assert.NoError(t, err)
			oo := OsOpSpec{
				Rename: &OsOpRenameSpec{
					NewPath: ValOrRef{Val: fmt.Sprintf("%s/%s", t.TempDir(), "newname.txt")},
					OldPath: ValOrRef{Val: f.Name()},
				},
			}
			assert.NoError(t, oo.Do(newMockActBuilder().testLogger(t).build()))
		})
		t.Run("should fail on non-existing file", func(t *testing.T) {
			oo := OsOpSpec{
				Rename: &OsOpRenameSpec{
					NewPath: ValOrRef{Val: "/////"},
					OldPath: ValOrRef{Val: "?????"},
				},
			}
			assert.Error(t, oo.Do(newMockActBuilder().testLogger(t).build()))
		})
	})
	t.Run("touch", func(t *testing.T) {
		t.Run("should update mtime of a file", func(t *testing.T) {
			d := t.TempDir()
			fi1, err := os.Stat(d)
			assert.NoError(t, err)
			assert.NotNil(t, fi1)
			time.Sleep(time.Millisecond * 10)
			oo := OsOpSpec{Touch: &OsOpTouchSpec{
				Path: ValOrRef{Val: d},
			}}
			assert.NoError(t, oo.Do(newMockActBuilder().testLogger(t).build()))
			fi2, err := os.Stat(d)
			assert.NoError(t, err)
			assert.NotNil(t, fi2)
			assert.Greater(t, fi2.ModTime().UnixNano(), fi1.ModTime().UnixNano())
		})
		t.Run("should fail on invalid path", func(t *testing.T) {
			oo := OsOpSpec{Touch: &OsOpTouchSpec{
				Path: ValOrRef{Val: "????////"},
			}}
			assert.Error(t, oo.Do(newMockActBuilder().testLogger(t).build()))
		})
	})

	t.Run("remove", func(t *testing.T) {
		for _, rec := range []bool{true, false} {
			t.Run(fmt.Sprintf("remove of temp dir should pass rec=%v", rec), func(t *testing.T) {
				oo := OsOpSpec{Remove: &OsOpRemoveSpec{Recursive: &rec, Path: ValOrRef{Val: t.TempDir()}}}
				t.Log(oo.String())
				assert.NoError(t, oo.Do(mockEmptyActCtx()))
			})
		}
		t.Run("remove of non-existent directory should fail", func(t *testing.T) {
			oo := OsOpSpec{Remove: &OsOpRemoveSpec{Path: ValOrRef{Val: filepath.Join(t.TempDir(), "098765432")}}}
			t.Log(oo.String())
			assert.Error(t, oo.Do(mockEmptyActCtx()))
		})
	})
	t.Run("link", func(t *testing.T) {
		for _, sym := range []bool{true, false} {
			t.Run(fmt.Sprintf("missing old/new file should fail (sym:%v)", sym), func(t *testing.T) {
				oo := OsOpSpec{Link: &OsOpLinkSpec{Symbolic: sym}}
				t.Log(oo.String())
				assert.Error(t, oo.Do(mockEmptyActCtx()))
			})
		}
		t.Run("link to existing file should pass", func(t *testing.T) {
			oldPath := filepath.Join(t.TempDir(), "old")
			newPath := filepath.Join(t.TempDir(), "new")
			assert.NoError(t, os.WriteFile(oldPath, []byte("old"), 0777))
			oo := OsOpSpec{Link: &OsOpLinkSpec{
				OldName:  ValOrRef{Val: oldPath},
				NewName:  ValOrRef{Val: newPath},
				Symbolic: true,
			}}
			t.Log(oo.String())
			assert.NoError(t, oo.Do(mockEmptyActCtx()))
		})
	})
	t.Run("copy", func(t *testing.T) {
		t.Run("copy from one dir to another should work", func(t *testing.T) {
			from := t.TempDir()
			to := t.TempDir()
			assert.NoError(t, os.WriteFile(filepath.Join(from, "file1.txt"), []byte("aaa"), 0777))
			oo := OsOpSpec{Copy: &OsOpCopySpec{
				From: ValOrRef{Val: from},
				To:   ValOrRef{Val: to},
			}}
			t.Log(oo.String())
			assert.NoError(t, oo.Do(mockEmptyActCtx()))
		})

		t.Run("copy from non-existent dir should fail", func(t *testing.T) {
			from := t.TempDir()
			to := t.TempDir()
			assert.NoError(t, os.WriteFile(filepath.Join(from, "file1.txt"), []byte("aaa"), 0777))
			oo := OsOpSpec{Copy: &OsOpCopySpec{
				From: ValOrRef{Val: filepath.Join(from, "this does not exist")},
				To:   ValOrRef{Val: to},
			}}
			t.Log(oo.String())
			assert.Error(t, oo.Do(mockEmptyActCtx()))
		})
	})
}
