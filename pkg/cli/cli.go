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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gookit/color"
	xlog "github.com/rkosegi/slog-config"
	ytp "github.com/rkosegi/yaml-pipeline/pkg/pipeline"
	"github.com/rkosegi/yaml-pipeline/pkg/utils"
	"github.com/rkosegi/yaml-pipeline/pkg/version"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
	"gopkg.in/yaml.v3"
)

type data struct {
	sc       *xlog.SlogConfig
	file     string
	validate bool
	logger   *slog.Logger
	pp       ytp.PipelineSpec
	vals     []string
}

var (
	startOpStyle = color.Style{color.FgMagenta}
	endOpStyle   = color.Style{color.FgGreen}
	errStyle     = color.Style{color.FgRed, color.OpBold}
)

type simpleListener struct {
	ind int
	l   *slog.Logger
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

func preRun(d *data) error {
	if d.validate {
		fmt.Fprintln(os.Stderr, color.Blue.Render(fmt.Sprintf("[Schema] Validating document")))
		res, err := utils.ValidateFileAgainstSchema(d.file)
		if err != nil {
			return err
		}
		if res != nil && !res.Valid {
			tree := treeprint.New()
			utils.DumpSchemaEvalResultToTree(tree, res.Details)
			fmt.Fprintln(os.Stderr, tree.String())
			return res
		}
		fmt.Fprintln(os.Stderr, color.Blue.Render(fmt.Sprintf("[Schema] OK")))
	}
	return nil
}

func setValues(d *data, gd dom.ContainerBuilder) {
	fmt.Fprintln(os.Stderr, color.Blue.Render(fmt.Sprintf("[Values] Setting values")))
	utils.ApplyVarsToDom(d.pp.Vars, "vars", gd)
	utils.ApplyValues(gd, d.vals)
	fmt.Fprintln(os.Stderr, color.Blue.Render(fmt.Sprintf("[Values] OK")))
}

func run(d *data) error {
	var (
		bytes []byte
		err   error
	)
	if bytes, err = os.ReadFile(d.file); err != nil {
		return err
	}
	if err = yaml.Unmarshal(bytes, &d.pp); err != nil {
		return err
	}
	gd := dom.ContainerNode()
	setValues(d, gd)
	return ytp.New(
		ytp.WithData(gd),
		ytp.WithListener(&simpleListener{l: d.logger}),
	).Execute(d.pp.ActionSpec)
}

func New() *cobra.Command {
	d := &data{
		sc:       xlog.MustNew("info", xlog.LogFormatLogFmt),
		validate: true,
	}

	short := "Runs a pipeline from a file"
	cmd := &cobra.Command{
		Use:   "yp",
		Short: short,
		Long: short + "\n" + `
File is validated against JSON schema unless validation is explicitly disabled (--validate false).
Initial values can be set using --set keyX=valueY.


`,
		Version: version.Get(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			d.logger = d.sc.Logger()
			if len(d.file) == 0 {
				return errors.New("pipeline file is required (--file)")
			}
			return preRun(d)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(d)
		},
	}
	cmd.Flags().BoolVar(&d.validate, "validate", d.validate,
		"Whether to validate pipeline file against current JSON schema")
	cmd.Flags().StringVar(&d.file, "file", "", "pipeline file to run")
	cmd.Flags().StringArrayVar(&d.vals, "set", d.vals, "set value to data tree prior to run")
	return cmd
}
