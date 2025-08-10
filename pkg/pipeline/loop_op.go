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

import "reflect"

func (l *LoopOpSpec) String() string {
	return "Loop[]"
}

func (l *LoopOpSpec) doAction(ctx ActionContext, act Action) (err error) {
	if act != nil && !reflect.ValueOf(act).IsNil() {
		return ctx.Executor().Execute(act)
	}
	return nil
}

func (l *LoopOpSpec) Do(ctx ActionContext) (err error) {
	if err = l.doAction(ctx, l.Init); err != nil {
		return err
	}

	for {
		ctx.InvalidateSnapshot()
		var next bool
		if next, err = ctx.TemplateEngine().EvalBool(l.Test, ctx.Snapshot()); err != nil {
			return err
		}
		if next {
			if err = l.doAction(ctx, l.PostAction); err != nil {
				return err
			}
			if err = l.Action.Do(ctx); err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}

func (l *LoopOpSpec) CloneWith(ctx ActionContext) Action {
	lc := new(LoopOpSpec)
	lc.Test = l.Test
	lc.Action = l.Action.CloneWith(ctx).(ActionSpec)
	if l.Init != nil {
		lc.Init = ptr(l.Init.CloneWith(ctx).(ActionSpec))
	}
	if l.PostAction != nil {
		lc.PostAction = ptr(l.PostAction.CloneWith(ctx).(ActionSpec))
	}
	return lc
}
