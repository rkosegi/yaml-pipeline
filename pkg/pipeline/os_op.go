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
	"time"

	cp "github.com/otiai10/copy"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/samber/lo"
)

func (m *OsOpMkdirSpec) String() string {
	return fmt.Sprintf("mkdir[%v]", m.Path.String())
}

func (m *OsOpChdirSpec) String() string {
	return fmt.Sprintf("chdir[%v]", m.Path.String())
}

func (m *OsOpChmodSpec) String() string {
	return fmt.Sprintf("chmod[%v]", m.Path.String())
}

func (m *OsOpGetcwdSpec) String() string {
	return fmt.Sprintf("getcwd[storeTo=%v]", m.StoreTo)
}

func (m *OsOpHostnameSpec) String() string {
	return fmt.Sprintf("hostname[storeTo=%v]", m.StoreTo)
}

func (m *OsOpUserHomeSpec) String() string {
	return fmt.Sprintf("userhome[storeTo=%v]", m.StoreTo)
}

func (m *OsOpCopySpec) String() string {
	return fmt.Sprintf("copy[from=%s,to=%s]", m.From.String(), m.To.String())
}

func (m *OsOpRemoveSpec) String() string {
	return fmt.Sprintf("remove[recursive=%v, path=%s]", safeBoolDeref(m.Recursive), m.Path.String())
}

func (m *OsOpStatSpec) String() string {
	return fmt.Sprintf("stat[path=%s]", m.Path.String())
}

func (m *OsOpReadDirSpec) String() string {
	return fmt.Sprintf("readdir[path=%s]", m.Path.String())
}

func (m *OsOpLinkSpec) String() string {
	return fmt.Sprintf("link[old=%s,new=%s]", m.OldName.String(), m.NewName.String())
}

func fileInfoToDom(fi os.FileInfo) dom.Container {
	return dom.ContainerNode().
		AddValue("mode", dom.LeafNode(fi.Mode())).
		AddValue("size", dom.LeafNode(fi.Size())).
		AddValue("mtime", dom.LeafNode(fi.ModTime()))
}

func dirEntryToDom(ent os.DirEntry) dom.Container {
	x := dom.ContainerNode().
		AddValue("name", dom.LeafNode(ent.Name())).
		AddValue("type", dom.LeafNode(uint32(ent.Type()))).
		AddValue("isDir", dom.LeafNode(ent.IsDir()))
	if fi, _ := ent.Info(); fi != nil {
		x.AddValue("info", fileInfoToDom(fi))
	}
	return x
}

func (o OsOpSpec) Do(ctx ActionContext) error {
	var err error
	if o.Mkdir != nil {
		var fn = os.Mkdir
		m := os.FileMode(0o775)
		if o.Mkdir.Mode != nil {
			m = *o.Mkdir.Mode
		}
		if safeBoolDeref(o.Mkdir.Recursive) {
			fn = os.MkdirAll
		}
		p := o.Mkdir.Path.Resolve(ctx)
		ctx.Logger().Log(fmt.Sprintf("mkdir: creating directory %s", p))
		if err = fn(p, m); err != nil {
			return err
		}
	}
	if o.Chdir != nil {
		p := o.Chdir.Path.Resolve(ctx)
		ctx.Logger().Log(fmt.Sprintf("chdir: changing directory to %s", p))
		if err = os.Chdir(p); err != nil {
			return err
		}
	}
	if o.Chmod != nil {
		p := o.Chmod.Path.Resolve(ctx)
		ctx.Logger().Log(fmt.Sprintf("chmod: changing mode of %s to %o", p, o.Chmod.Mode))
		if err = os.Chmod(p, o.Chmod.Mode); err != nil {
			return err
		}
	}
	if o.Getcwd != nil {
		d, _ := os.Getwd()
		ctx.Data().Set(pp.MustParse(o.Getcwd.StoreTo.Resolve(ctx)), dom.LeafNode(d))
	}
	if o.Hostname != nil {
		d, _ := os.Hostname()
		ctx.Data().Set(pp.MustParse(o.Hostname.StoreTo.Resolve(ctx)), dom.LeafNode(d))
	}
	if o.Link != nil {
		var fn = os.Link
		op := o.Link.OldName.Resolve(ctx)
		np := o.Link.NewName.Resolve(ctx)
		ctx.Logger().Log(fmt.Sprintf("link: creating link from %s to %s", op, np))
		if o.Link.Symbolic {
			fn = os.Symlink
		}
		if err = fn(op, np); err != nil {
			return err
		}
	}
	if o.Remove != nil {
		var fn = os.Remove
		if safeBoolDeref(o.Remove.Recursive) {
			fn = os.RemoveAll
		}
		p := o.Remove.Path.Resolve(ctx)
		ctx.Logger().Log(fmt.Sprintf("remove: removing path %s", p))
		if err = fn(o.Remove.Path.Resolve(ctx)); err != nil {
			return err
		}
	}
	if o.Rename != nil {
		op := o.Rename.OldPath.Resolve(ctx)
		np := o.Rename.NewPath.Resolve(ctx)
		ctx.Logger().Log(fmt.Sprintf("rename: renaming %s => %s", op, np))
		if err = os.Rename(op, np); err != nil {
			return err
		}
	}
	if o.Touch != nil {
		p := o.Touch.Path.Resolve(ctx)
		now := time.Now()
		ctx.Logger().Log(fmt.Sprintf("touch: updating atime/mtime of %s to %s", p, now.String()))
		if err = os.Chtimes(p, now, now); err != nil {
			return err
		}
	}
	if o.Userhome != nil {
		var d string
		if d, err = os.UserHomeDir(); err != nil {
			return err
		}
		ctx.Data().Set(pp.MustParse(o.Userhome.StoreTo.Resolve(ctx)), dom.LeafNode(d))
	}
	if o.Readdir != nil {
		var de []os.DirEntry
		if de, err = os.ReadDir(o.Readdir.Path.Resolve(ctx)); err != nil {
			return err
		}
		var entries = lo.Map(de, func(item os.DirEntry, _ int) dom.Node {
			return dirEntryToDom(item)
		})
		ctx.Data().Set(pp.MustParse(o.Readdir.StoreTo.Resolve(ctx)), dom.ListNode(entries...))
	}
	if o.Stat != nil {
		var fi os.FileInfo
		if fi, err = os.Stat(o.Stat.Path.Resolve(ctx)); err != nil {
			return err
		}
		ctx.Data().Set(pp.MustParse(o.Stat.StoreTo.Resolve(ctx)), fileInfoToDom(fi))
	}
	if o.Copy != nil {
		from := o.Copy.From.Resolve(ctx)
		to := o.Copy.To.Resolve(ctx)
		ctx.Logger().Log(fmt.Sprintf("copy: copying file(s) from %s to %s", from, to))
		if err = cp.Copy(from, to); err != nil {
			return err
		}
	}

	return nil
}

func (o OsOpSpec) String() string {
	return fieldStringer("OS", o)
}

func (o OsOpSpec) CloneWith(ctx ActionContext) Action {
	return cloneFieldsWith[OsOpSpec](o, ctx)
}
