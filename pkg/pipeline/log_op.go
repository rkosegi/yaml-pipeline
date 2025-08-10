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

func (lo *LogOpSpec) Do(ctx ActionContext) error {
	ctx.Logger().Log(ctx.TemplateEngine().RenderLenient(lo.Message, ctx.Snapshot()))
	return nil
}

func (lo *LogOpSpec) String() string {
	return fmt.Sprintf("Log[message(%d)=%s]", len(lo.Message), strTruncIfNeeded(lo.Message, 25))
}

func (lo *LogOpSpec) CloneWith(ctx ActionContext) Action {
	return &LogOpSpec{
		Message: ctx.TemplateEngine().RenderLenient(lo.Message, ctx.Snapshot()),
	}
}
