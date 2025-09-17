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

	"github.com/rkosegi/yaml-toolkit/patch"
)

func (ps *PatchOpSpec) String() string {
	return fmt.Sprintf("Patch[Op=%s,Path=%s]", ps.Op, ps.Path)
}

func (ps *PatchOpSpec) Do(ctx ActionContext) error {
	var err error
	ss := ctx.Snapshot()
	oo := &patch.OpObj{
		Op: ps.Op,
	}
	var path patch.Path
	path, err = patch.ParsePath(ctx.TemplateEngine().RenderLenient(ps.Path, ss))
	if err != nil {
		return errWithInfo(err, "patch.ParsePath (path)")
	}
	oo.Path = path
	if ps.Value != nil {
		oo.Value = ps.Value.Value()
	} else if ps.ValueFrom != nil {
		oo.Value = ctx.Data().Get(pp.MustParse(ctx.TemplateEngine().RenderLenient(*ps.ValueFrom, ss)))
	}
	if !strIsEmpty(ps.From) {
		var from patch.Path
		from, err = patch.ParsePath(*ps.From)
		if err != nil {
			return errWithInfo(err, "patch.ParsePath (from)")
		}
		oo.From = &from
	}
	ctx.Logger().Log(fmt.Sprintf("Patch[Op=%v,Path=%v]", oo.Op, oo.Path))
	defer ctx.InvalidateSnapshot()
	return patch.Do(oo, ctx.Data())
}

func (ps *PatchOpSpec) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &PatchOpSpec{
		Op:        ps.Op,
		Value:     ps.Value,
		ValueFrom: safeRenderStrPointer(ps.ValueFrom, ctx.TemplateEngine(), ss),
		From:      safeRenderStrPointer(ps.From, ctx.TemplateEngine(), ss),
		Path:      ctx.TemplateEngine().RenderLenient(ps.Path, ss),
	}
}
