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
)

func (am ActionMeta) String() string {
	var (
		sb    strings.Builder
		parts []string
	)
	sb.WriteByte('[')
	if !strIsEmpty(am.Name) {
		parts = append(parts, fmt.Sprintf("name=%s", *am.Name))
	}
	if am.Order != nil {
		parts = append(parts, fmt.Sprintf("order=%d", *am.Order))
	}
	when := strings.TrimSpace(safeStrDeref(am.When))
	if len(when) > 0 {
		parts = append(parts, fmt.Sprintf("when=%s", when))
	}
	sb.WriteString(strings.Join(parts, ","))
	sb.WriteByte(']')
	return sb.String()
}
