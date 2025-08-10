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

import "fmt"

func (e *ExtOpSpec) String() string {
	return fmt.Sprintf("Ext[func=%s,args=%d]", e.Function, safeSize(e.Args))
}

func (e *ExtOpSpec) Do(ctx ActionContext) error {
	if e.Args == nil {
		e.Args = &map[string]interface{}{}
	}
	if fn := ctx.Ext().GetActionFactory(e.Function); fn != nil {
		return ctx.Executor().Execute(fn.ForArgs(ctx, *e.Args))
	}
	return fmt.Errorf("no such function: %s", e.Function)
}

func (e *ExtOpSpec) CloneWith(_ ActionContext) Action {
	return &ExtOpSpec{Function: e.Function}
}
