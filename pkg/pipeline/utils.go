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
	"reflect"
	"regexp"
	"slices"
	"strings"

	te "github.com/rkosegi/yaml-pipeline/pkg/pipeline/template_engine"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
)

var pp = props.NewPathParser()

func strTruncIfNeeded(in string, size int) string {
	if len(in) <= size {
		return in
	}
	return in[0:size]
}

func parseFile(path string, mode ParseFileMode) (dom.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(mode) == 0 {
		mode = ParseFileModeText
	}
	val, err := mode.toValue(data)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func safeCloneValOrRef(v *ValOrRef, ctx ActionContext) *ValOrRef {
	if v == nil {
		return nil
	}
	return v.CloneWith(ctx)
}

func safeStrDeref(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}

func setStrategyPointer(s SetStrategy) *SetStrategy {
	return &s
}

func safeRegexpDeref(re *regexp.Regexp) string {
	if re == nil {
		return ""
	}
	return re.String()
}

func safeSizeReflect(rv reflect.Value) int {
	if rv.Kind() == reflect.Ptr {
		return safeSizeReflect(rv.Elem())
	}
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array || rv.Kind() == reflect.Map {
		return rv.Len()
	}
	return 0
}

// checks for size of slice or map. Anything else is considered as size of 0.
func safeSize(in any) int {
	if in == nil {
		return 0
	}
	return safeSizeReflect(reflect.ValueOf(in))
}

func nonEmpty(in *string) bool {
	return in != nil && len(*in) > 0
}

func actionSpecFix(as ActionSpec) ActionSpec {
	if as.Order == nil {
		as.Order = ptr(0)
	}
	return as
}

func sortActionNames(actions ChildActions) []string {
	var keys []string
	for n := range actions {
		keys = append(keys, n)
	}
	slices.SortFunc(keys, func(a, b string) int {
		return *actionSpecFix(actions[a]).Order - *actionSpecFix(actions[b]).Order
	})
	return keys
}

func actionNames(actions ChildActions) string {
	return strings.Join(sortActionNames(actions), ",")
}

func safeCopyIntSlice(in *[]int) *[]int {
	if in == nil {
		return nil
	}
	r := make([]int, len(*in))
	copy(r, *in)
	return &r
}

func safeRenderStrPointer(str *string, teng te.TemplateEngine, data map[string]interface{}) *string {
	if str == nil {
		return nil
	}
	s := teng.RenderLenient(*str, data)
	return &s
}

func safeRenderStrSlice(args *[]string, teng te.TemplateEngine, data map[string]interface{}) *[]string {
	if args == nil {
		return nil
	}
	r := make([]string, len(*args))
	for i, arg := range *args {
		r[i] = teng.RenderLenient(arg, data)
	}
	return &r
}

func safeBoolDeref(in *bool) bool {
	if in == nil {
		return false
	}
	return *in
}

func ptr[T any](v T) *T {
	return &v
}

func fieldStringer(parent string, ptrVal any) string {
	var sb strings.Builder
	parts := make([]string, 0)
	sb.WriteString(parent)
	sb.WriteString("[")
	ptrType := reflect.TypeOf(ptrVal)
	asv := reflect.ValueOf(ptrVal)

	if ptrType.Kind() == reflect.Ptr {
		ptrType = ptrType.Elem()
		asv = asv.Elem()
	}
	fields := reflect.VisibleFields(ptrType)

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

func cloneFieldsWith[T Action](as Action, ctx ActionContext) Action {
	ret := new(T)
	actType := reflect.TypeOf(as)
	srcVal := reflect.ValueOf(as)
	dstVal := reflect.ValueOf(ret)
	fields := reflect.VisibleFields(actType)
	for _, field := range fields {
		srcField := srcVal.FieldByIndex(field.Index)
		if !reflect.ValueOf(srcField.Interface()).IsNil() {
			var out interface{}
			out = srcField.Interface()
			if cloneable, ok := out.(Cloneable); ok {
				out = cloneable.CloneWith(ctx)
			}
			dstField := dstVal.Elem().FieldByName(field.Name)
			dstField.Set(reflect.ValueOf(out))
		}
	}
	return *ret
}
