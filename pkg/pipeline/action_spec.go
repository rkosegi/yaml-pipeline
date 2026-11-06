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
)

type ActionSpec struct {
	ActionMeta `yaml:",inline" json:",inline"`
	// Operations to perform
	Operations OpSpec `yaml:",inline" json:",inline"`
	// Children element is an optional map of child actions that will be executed
	// as a part of this action (after any of OpSpec in Operations are performed).
	// Exact order of execution is given by Order field value (lower the value, sooner the execution will take place).
	Children ChildActions `yaml:"steps,omitempty" json:"steps,omitempty"`
}

func (s ActionSpec) CloneWith(ctx ActionContext) Action {
	return ActionSpec{
		ActionMeta: s.ActionMeta,
		Operations: s.Operations.CloneWith(ctx).(OpSpec),
		Children:   s.Children.CloneWith(ctx).(ChildActions),
	}
}

func (s ActionSpec) String() string {
	return fmt.Sprintf("ActionSpec[meta=%v]", s.ActionMeta)
}

func (s ActionSpec) shouldPropagateError() bool {
	return s.ErrorPropagation == nil || *s.ErrorPropagation != ErrorPropagationPolicyIgnore
}

func (s ActionSpec) Do(ctx ActionContext) error {
	for _, a := range []Action{s.Operations, s.Children} {
		if s.When != nil {
			if ok, err := ctx.TemplateEngine().EvalBool(*s.When, ctx.Snapshot()); err != nil {
				return errWithInfo(err, "action/evalBool")
			} else if !ok {
				ctx.Logger().Log("tag::skip", "execution skipped as condition evaluated to false =>",
					safeStrTrim(s.When))
				return nil
			}
		}
		if err := s.execute(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (s ActionSpec) execute(ctx ActionContext, act Action) error {
	if s.Lifecycle == nil {
		s.Lifecycle = &ActionLifecycle{}
	}
	if s.Lifecycle.OnComplete != nil {
		defer func() {
			_ = ctx.Executor().ExecuteContext(ctx.Context(), s.Lifecycle.OnComplete)
		}()
	}

	if err := ctx.Executor().ExecuteContext(ctx.Context(), act); err != nil {
		if s.Lifecycle.OnFailure != nil {
			defer func() {
				_ = ctx.Executor().ExecuteContext(ctx.Context(), s.Lifecycle.OnFailure)
			}()
		}
		if s.shouldPropagateError() {
			return err
		}
		return nil
	} else {
		if s.Lifecycle.OnSuccess != nil {
			_ = ctx.Executor().ExecuteContext(ctx.Context(), s.Lifecycle.OnSuccess)
		}
	}
	return nil
}
