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
)

func (m *OsOpMkdirSpec) String() string {
	return fmt.Sprintf("mkdir[%s]", m.Path.String())
}
func (m *OsOpChdirSpec) String() string {
	return fmt.Sprintf("chdir[%s]", m.Path.String())
}
func (m *OsOpChmodSpec) String() string {
	return fmt.Sprintf("chmod[%s]", m.Path.String())
}

func (o OsOpSpec) Do(ctx ActionContext) error {
	var err error
	if o.Mkdir != nil {
		m := os.FileMode(0o775)
		if o.Mkdir.Mode != nil {
			m = *o.Mkdir.Mode
		}
		p := o.Mkdir.Path.Resolve(ctx)
		ctx.Logger().Log(fmt.Sprintf("mkdir: creating directory %s", p))
		if err = os.Mkdir(p, m); err != nil {
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
	return nil
}

func (o OsOpSpec) String() string {
	return fieldStringer("OS", o)
}

func (o OsOpSpec) CloneWith(ctx ActionContext) Action {
	return cloneFieldsWith[OsOpSpec](o, ctx)
}
