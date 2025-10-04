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
	"errors"
	"fmt"
)

func (s *SwitchOpSpec) String() string {
	return fmt.Sprintf("switch[val=%v,cases=%d]", s.Expr, len(s.Cases))
}

func (s *SwitchOpSpec) CloneWith(ctx ActionContext) Action {
	return &SwitchOpSpec{
		Cases:   s.Cases.CloneWith(ctx).(ChildActions),
		Default: safeCloneActionSpec(s.Default, ctx),
		Expr:    *(s.Expr.CloneWith(ctx)),
	}
}

func (s *SwitchOpSpec) Do(ctx ActionContext) error {
	expr := s.Expr.Resolve(ctx)
	for k, v := range s.Cases {
		if value := ctx.TemplateEngine().RenderLenient(k, ctx.Snapshot()); value == expr {
			return v.Do(ctx)
		}
	}
	if s.Default != nil {
		return s.Default.Do(ctx)
	}
	return errors.New("neither the cases nor the default action is defined")
}
