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
)

func (tfo *TemplateFileOpSpec) String() string {
	return fmt.Sprintf("TemplateFile[File=%s,Output=%s]", tfo.File, tfo.Output)
}

func (tfo *TemplateFileOpSpec) Do(ctx ActionContext) error {
	if len(tfo.File) == 0 {
		return ErrFileEmpty
	}
	if len(tfo.Output) == 0 {
		return ErrOutputEmpty
	}
	ss := ctx.Snapshot()
	data := ctx.Data().AsContainer()
	p := "root"
	if tfo.Path != nil {
		p = ctx.TemplateEngine().RenderLenient(*tfo.Path, ss)
		if n := ctx.Data().Get(pp.MustParse(p)); n != nil && n.IsContainer() {
			data = n.AsContainer()
		} else {
			return fmt.Errorf("path does not point to a container: %s", *tfo.Path)
		}
	}
	inFile := ctx.TemplateEngine().RenderLenient(tfo.File, ss)
	ctx.Logger().Log(fmt.Sprintf("reading template file '%s' with data from %s", inFile, p))
	tmpl, err := os.ReadFile(inFile)
	if err != nil {
		return err
	}
	val, err := ctx.TemplateEngine().Render(string(tmpl), data.AsAny().(map[string]interface{}))
	if err != nil {
		return err
	}
	outFile := ctx.TemplateEngine().RenderLenient(tfo.Output, ss)
	ctx.Logger().Log(fmt.Sprintf("writing rendered template to '%s'", outFile))
	return os.WriteFile(outFile, []byte(val), 0644)
}

func (tfo *TemplateFileOpSpec) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &TemplateFileOpSpec{
		File:   ctx.TemplateEngine().RenderLenient(tfo.File, ss),
		Output: ctx.TemplateEngine().RenderLenient(tfo.Output, ss),
	}
}
