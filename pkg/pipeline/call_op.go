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

	"github.com/rkosegi/yaml-toolkit/dom"
)

func (c *CallOpSpec) String() string {
	return fmt.Sprintf("Call[Name=%s, Args=%d]", c.Name, safeSize(c.Args))
}

func (c *CallOpSpec) Do(ctx ActionContext) error {
	snap := ctx.Snapshot()
	ap := "args"
	if c.ArgsPath != nil {
		ap = *c.ArgsPath
	}
	ap = ctx.TemplateEngine().RenderLenient(ap, snap)
	if spec, exists := ctx.Ext().GetAction(c.Name); !exists {
		return fmt.Errorf("callable '%s' is not registered", c.Name)
	} else {
		ctx.Data().Set(pp.MustParse(ap), dom.DecodeAnyToNode(
			ctx.TemplateEngine().RenderMapLenient(*c.Args, snap)),
		)
		ctx.InvalidateSnapshot()
		defer func() {
			ctx.Data().Remove(ap)
			ctx.InvalidateSnapshot()
		}()
		return ctx.Executor().Execute(spec)
	}
}

func (c *CallOpSpec) CloneWith(_ ActionContext) Action {
	return &CallOpSpec{
		Name:     c.Name,
		Args:     c.Args,
		ArgsPath: c.ArgsPath,
	}
}
