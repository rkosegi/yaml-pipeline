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
	"errors"
	"fmt"
	"io"
	"os"
	osx "os/exec"
	"slices"
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
)

func (e *ExecOpSpec) String() string {
	return fmt.Sprintf("Exec[Program=%s,Dir=%s,Args=%d]", e.Program, safeStrDeref(e.Dir), safeSize(e.Args))
}

func (e *ExecOpSpec) Do(ctx ActionContext) error {
	var closables []io.Closer
	if e.ValidExitCodes == nil {
		e.ValidExitCodes = &[]int{}
	}
	if e.Args == nil {
		e.Args = &[]string{}
	}
	if e.Dir == nil {
		e.Dir = ptr("")
	}
	snapshot := ctx.Snapshot()
	prog := ctx.TemplateEngine().RenderLenient(e.Program, snapshot)
	args := *safeRenderStrSlice(e.Args, ctx.TemplateEngine(), snapshot)
	dir := safeRenderStrPointer(e.Dir, ctx.TemplateEngine(), snapshot)
	cmd := osx.Command(prog, args...)
	cmd.Dir = *dir
	defer func() {
		for _, closer := range closables {
			_ = closer.Close()
		}
	}()
	type streamTgt struct {
		output *string
		target *io.Writer
	}
	for _, stream := range []streamTgt{
		{
			output: e.Stdout,
			target: &cmd.Stdout,
		},
		{
			output: e.Stderr,
			target: &cmd.Stderr,
		},
	} {
		if stream.output != nil {
			s := ctx.TemplateEngine().RenderLenient(*stream.output, snapshot)
			stream.output = &s
			out, err := os.OpenFile(*stream.output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			*stream.target = out
			closables = append(closables, out)
		}
	}
	ctx.Logger().Log(fmt.Sprintf("prog=%s,dir=%s,args=[%s]", prog, safeStrDeref(dir), strings.Join(args, " ")))
	err := cmd.Run()
	var exitErr *osx.ExitError
	if errors.As(err, &exitErr) {
		if e.SaveExitCodeTo != nil {
			ctx.Data().AddValueAt(*e.SaveExitCodeTo, dom.LeafNode(exitErr.ExitCode()))
			ctx.InvalidateSnapshot()
		}
		if !slices.Contains(*e.ValidExitCodes, exitErr.ExitCode()) {
			return err
		}
	} else {
		return err
	}
	return nil
}

func (e *ExecOpSpec) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &ExecOpSpec{
		Program:        ctx.TemplateEngine().RenderLenient(e.Program, ss),
		Args:           safeRenderStrSlice(e.Args, ctx.TemplateEngine(), ss),
		Dir:            safeRenderStrPointer(e.Dir, ctx.TemplateEngine(), ss),
		Stdout:         safeRenderStrPointer(e.Stdout, ctx.TemplateEngine(), ss),
		Stderr:         safeRenderStrPointer(e.Stderr, ctx.TemplateEngine(), ss),
		ValidExitCodes: safeCopyIntSlice(e.ValidExitCodes),
	}
}
