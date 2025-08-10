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
	"reflect"
	"strings"
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
	ret := &OpSpec{}
	opSpecType := reflect.TypeOf(as)
	srcVal := reflect.ValueOf(as)
	dstVal := reflect.ValueOf(ret)
	fields := reflect.VisibleFields(opSpecType)
	for _, field := range fields {
		srcField := srcVal.FieldByIndex(field.Index)
		if !reflect.ValueOf(srcField.Interface()).IsNil() {
			cloned := srcField.Interface().(Action).CloneWith(ctx)
			dstField := dstVal.Elem().FieldByName(field.Name)
			dstField.Set(reflect.ValueOf(cloned))
		}
	}
	return *ret
}

func (as OpSpec) String() string {
	var sb strings.Builder
	parts := make([]string, 0)
	sb.WriteString("OpSpec[")

	asv := reflect.ValueOf(as)
	fields := reflect.VisibleFields(reflect.TypeOf(as))
	for _, field := range fields {
		x := asv.FieldByName(field.Name).Interface()
		if !reflect.ValueOf(x).IsNil() {
			parts = append(parts, fmt.Sprintf("%s=%v", field.Name, x.(fmt.Stringer).String()))
		}
	}

	sb.WriteString(strings.Join(parts, ","))
	sb.WriteString("]")
	return sb.String()
}
