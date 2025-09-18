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

package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/google/uuid"
	"github.com/gookit/color"
	ytp "github.com/rkosegi/yaml-pipeline/pkg/pipeline"
	"github.com/rkosegi/yaml-pipeline/pkg/utils"
)

var ol = map[string]ytp.Listener{
	"default": &simpleListener{},
	"gitlab": &gitlabListener{
		simpleListener: &simpleListener{},
		sec:            arraystack.New(),
	},
}

// https://docs.gitlab.com/ci/jobs/job_logs/#use-a-script-to-improve-display-of-collapsible-sections
type gitlabListener struct {
	sec *arraystack.Stack
	*simpleListener
}

type simpleListener struct {
	ind int
}

func (g *gitlabListener) OnBefore(ctx ytp.ActionContext) {
	if as, ok := ctx.Action().(ytp.ActionSpec); ok {
		secId := uuid.New().String()
		fmt.Fprintf(os.Stderr, "\x1b[0Ksection_start:%d:action_spec_%s[collapsed=true]\r\x1b[0K%s\n", time.Now().Unix(),
			secId, as.String())
		g.sec.Push(secId)
	} else {
		g.simpleListener.OnBefore(ctx)
	}
}

func (g *gitlabListener) OnAfter(ctx ytp.ActionContext, err error) {
	if _, ok := ctx.Action().(ytp.ActionSpec); ok {
		secId, _ := g.sec.Pop()
		fmt.Fprintf(os.Stderr, "\x1b[0Ksection_end:%d:action_spec_%s\r\x1b[0K\n", time.Now().Unix(), secId.(string))
	} else {
		g.simpleListener.OnAfter(ctx, err)
	}
}

func (g *gitlabListener) OnLog(ctx ytp.ActionContext, v ...interface{}) {
	g.simpleListener.OnLog(ctx, v...)
}

func (s *simpleListener) indentStr() string {
	return strings.Repeat(" ", s.ind)
}

func (s *simpleListener) OnBefore(ctx ytp.ActionContext) {
	fmt.Fprintf(os.Stderr, "%s %s %v\n", startOpStyle.Render("[Start]"), s.indentStr(), ctx.Action())
	s.ind++
}

func (s *simpleListener) OnAfter(ctx ytp.ActionContext, err error) {
	s.ind -= 1
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s %v\n", errStyle.Render("[Error]"), s.indentStr(), ctx.Action())
		panic(err)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s %v\n", endOpStyle.Render("[Done ]"), s.indentStr(), ctx.Action())
	}
}

func (s *simpleListener) OnLog(_ ytp.ActionContext, v ...interface{}) {
	if tag, hasTag := utils.GetLogTag(v...); hasTag {
		switch tag {
		case "skip":
			fmt.Fprint(os.Stderr, color.Gray.Render(fmt.Sprintf("[SKIP ] %s %v\n", s.indentStr(), v[1:])))
			return
		}
	}
	fmt.Fprint(os.Stderr, color.Blue.Render(fmt.Sprintf("[Log  ] %s %v\n", s.indentStr(), v)))
}
