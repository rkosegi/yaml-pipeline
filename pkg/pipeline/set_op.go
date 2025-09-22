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
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
)

func strIsEmpty(s *string) bool {
	return s == nil || len(strings.TrimSpace(*s)) == 0
}

type setHandlerFn func(path *string, orig, other dom.ContainerBuilder)

func setOpMergeIfContainersReplaceOtherwise(orig, other dom.ContainerBuilder) {
	for k, v := range other.Children() {
		origChild := orig.Child(k)
		if origChild != nil && origChild.IsContainer() && v.IsContainer() {
			orig.AddValue(k, origChild.(dom.ContainerBuilder).Merge(v.AsContainer()))
		} else {
			orig.AddValue(k, v)
		}
	}
}

var setHandlerFnMap = map[SetStrategy]setHandlerFn{
	SetStrategyMerge: func(path *string, orig, other dom.ContainerBuilder) {
		if !strIsEmpty(path) {
			dest := orig.Get(pp.MustParse(*path))
			if dest != nil && dest.IsContainer() {
				orig.Set(pp.MustParse(*path), dest.(dom.ContainerBuilder).Merge(other))
			} else {
				orig.Set(pp.MustParse(*path), other)
			}
		} else {
			setOpMergeIfContainersReplaceOtherwise(orig, other)
		}
	},
	SetStrategyReplace: func(path *string, orig, other dom.ContainerBuilder) {
		if !strIsEmpty(path) {
			orig.Set(pp.MustParse(*path), other)
		} else {
			for k, v := range other.Children() {
				orig.Set(pp.MustParse(k), v)
			}
		}
	},
}

func (sa *SetOpSpec) String() string {
	return fmt.Sprintf("Set[Path=%s]", safeStrDeref(sa.Path))
}

func (sa *SetOpSpec) Do(ctx ActionContext) error {
	gd := ctx.Data()
	if sa.Data == nil {
		return ErrNoDataToSet
	}
	if sa.Strategy == nil {
		sa.Strategy = setStrategyPointer(SetStrategyMerge)
	}
	handler, exists := setHandlerFnMap[*sa.Strategy]
	if !exists {
		return fmt.Errorf("SetOpSpec: unknown SetStrategy %s", *sa.Strategy)
	}
	indata := sa.Data
	if safeBoolDeref(sa.Render) {
		indata = ctx.TemplateEngine().RenderMapLenient(sa.Data, ctx.Snapshot())
	}
	data := dom.DecodeAnyToNode(indata).(dom.ContainerBuilder)
	handler(sa.Path, gd, data)
	ctx.InvalidateSnapshot()
	return nil
}

func (sa *SetOpSpec) CloneWith(ctx ActionContext) Action {
	return &SetOpSpec{
		Data:     sa.Data,
		Path:     safeRenderStrPointer(sa.Path, ctx.TemplateEngine(), ctx.Snapshot()),
		Strategy: sa.Strategy,
	}
}
