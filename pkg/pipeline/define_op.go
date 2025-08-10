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

func (d *DefineOpSpec) String() string {
	return fmt.Sprintf("Define[Name=%s, Action=%v]", d.Name, d.Action)
}

func (d *DefineOpSpec) Do(ctx ActionContext) error {
	if _, exists := ctx.Ext().GetAction(d.Name); exists {
		return fmt.Errorf("callable '%s' is already defined", d.Name)
	}
	ctx.Ext().DefineAction(d.Name, d.Action)
	return nil
}

func (d *DefineOpSpec) CloneWith(ctx ActionContext) Action {
	return &DefineOpSpec{
		Name:   d.Name,
		Action: d.Action.CloneWith(ctx).(ActionSpec),
	}
}
