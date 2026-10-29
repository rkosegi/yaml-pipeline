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

	osx "os/exec"
	"slices"
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
)

func (e *ExecOpSpec) String() string {
	return fmt.Sprintf("Exec[Program=%s,Dir=%s,Args=%d]", e.Program, safeStrDeref(e.Dir), safeSize(e.Args))
}

func (e *ExecOpSpec) ensureValues() {
	if e.ValidExitCodes == nil {
		e.ValidExitCodes = &[]int{}
	}
	if e.Args == nil {
		e.Args = &[]string{}
	}
	if e.Dir == nil {
		e.Dir = new("")
	}
}

func (e *ExecOpSpec) Do(ctx ActionContext) error {
	var closables []io.Closer
	e.ensureValues()
	snapshot := ctx.Snapshot()
	prog := ctx.TemplateEngine().RenderLenient(e.Program, snapshot)
	args := *safeRenderStrSlice(e.Args, ctx.TemplateEngine(), snapshot)
	dir := safeRenderStrPointer(e.Dir, ctx.TemplateEngine(), snapshot)
	cmd := osx.CommandContext(ctx.Context(), prog, args...)
	cmd.Dir = *dir
	defer func() {
		for _, closer := range closables {
			_ = closer.Close()
		}
	}()
	type streamTgt struct {
		sd     *StreamDestination
		output *string
		target *io.Writer
	}
	for _, stream := range []streamTgt{
		{
			sd: e.Stdout,
			// output: e.Stdout,
			target: &cmd.Stdout,
		},
		{
			sd: e.Stderr,
			// output: e.Stderr,
			target: &cmd.Stderr,
		},
	} {
		if stream.sd == nil {
			stream.sd = &StreamDestination{}
		}
		if err := stream.sd.Resolve(ctx); err != nil {
			return err
		}
		closables = append(closables, stream.sd)
	}
	ctx.Logger().Log(fmt.Sprintf("prog=%s,dir=%s,args=[%s]", prog, safeStrDeref(dir), strings.Join(args, " ")))
	err := cmd.Run()
	if exitErr, ok := errors.AsType[*osx.ExitError](err); ok {
		if e.SaveExitCodeTo != nil {
			ctx.Data().Set(pp.MustParse(*e.SaveExitCodeTo), dom.LeafNode(exitErr.ExitCode()))
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
		Stdout:         safeCloneStreamDestination(e.Stdout, ctx),
		Stderr:         safeCloneStreamDestination(e.Stderr, ctx),
		ValidExitCodes: safeCopyIntSlice(e.ValidExitCodes),
	}
}
