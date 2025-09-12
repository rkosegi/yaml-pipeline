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
	"reflect"
)

func (as OpSpec) IsEmpty() bool {
	return len(as.toList()) == 0
}

func (as OpSpec) toList() []Action {
	actions := make([]Action, 0)
	asv := reflect.ValueOf(as)
	fields := reflect.VisibleFields(reflect.TypeOf(as))
	for _, field := range fields {
		x := asv.FieldByName(field.Name).Interface()
		if !reflect.ValueOf(x).IsNil() {
			actions = append(actions, x.(Action))
		}
	}
	return actions
}

func (as OpSpec) Do(ctx ActionContext) error {
	for _, a := range as.toList() {
		err := ctx.Executor().Execute(a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (as OpSpec) CloneWith(ctx ActionContext) Action {
	return cloneFieldsWith[OpSpec](as, ctx)
}

func (as OpSpec) String() string {
	return fieldStringer("OpSpec", &as)
}
