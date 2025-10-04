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
	"strings"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
)

// for mock purposes only. this could be used to override os.Environ() to arbitrary func
var (
	envGetter = os.Environ
)

func (eo *EnvOpSpec) Do(ctx ActionContext) error {
	var (
		inclFn common.StringPredicateFn
		exclFn common.StringPredicateFn
	)
	p := ""
	if eo.Path != nil {
		p = *eo.Path
	}
	inclFn = common.MatchAny()
	exclFn = common.MatchNone()
	if eo.Include != nil {
		inclFn = common.MatchRe(eo.Include)
	}
	if eo.Exclude != nil {
		exclFn = common.MatchRe(eo.Exclude)
	}
	for _, env := range envGetter() {
		parts := strings.SplitN(env, "=", 2)
		if inclFn(parts[0]) && !exclFn(parts[0]) {
			k := prefixPath(p, fmt.Sprintf("Env.%s", parts[0]))
			ctx.Data().Set(pp.MustParse(k), dom.LeafNode(parts[1]))
		}
	}
	ctx.InvalidateSnapshot()
	return nil
}

func prefixPath(parent string, key string) string {
	return strings.TrimPrefix(parent+"."+key, ".")
}

func (eo *EnvOpSpec) String() string {
	return fmt.Sprintf("Env[Path=%s,incl=%s,excl=%s]", safeStrDeref(eo.Path),
		safeRegexpDeref(eo.Include), safeRegexpDeref(eo.Exclude))
}

func (eo *EnvOpSpec) CloneWith(ctx ActionContext) Action {
	return &EnvOpSpec{
		Include: eo.Include,
		Exclude: eo.Exclude,
		Path:    safeRenderStrPointer(eo.Path, ctx.TemplateEngine(), ctx.Snapshot()),
	}
}
