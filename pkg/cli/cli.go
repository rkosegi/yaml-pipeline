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
	"os"
	"strings"

	"github.com/gookit/color"
	xlog "github.com/rkosegi/slog-config"
	ytp "github.com/rkosegi/yaml-pipeline/pkg/pipeline"
	"github.com/rkosegi/yaml-pipeline/pkg/utils"
	"github.com/rkosegi/yaml-pipeline/pkg/version"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
	"gopkg.in/yaml.v3"
)

type data struct {
	sc       *xlog.SlogConfig
	file     string
	validate bool
	pp       ytp.PipelineSpec
	vals     []string
	output   string
	sl       ytp.Listener
	colorUse string
}

var (
	startOpStyle = color.Style{color.FgMagenta}
	endOpStyle   = color.Style{color.FgGreen}
	errStyle     = color.Style{color.FgRed, color.OpBold}
)

func preRun(d *data) error {
	applyColorUse(d.colorUse)
	if l, ok := ol[d.output]; ok {
		d.sl = l
	} else {
		return fmt.Errorf("no such output decoration: %s, known types : %v", d.output, strings.Join(lo.Keys(ol), ","))
	}
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

func applyColorUse(use string) {
	switch use {
	case "always":
		_ = os.Setenv("FORCE_COLOR", "true")
	case "never":
		color.Enable = false
		_ = os.Unsetenv("FORCE_COLOR")
	}
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
		ytp.WithListener(d.sl),
	).Execute(d.pp.ActionSpec)
}

func New() *cobra.Command {
	d := &data{
		validate: true,
		output:   "default",
		colorUse: "always",
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
			if len(d.file) == 0 {
				return errors.New("pipeline file is required (--file)")
			}
			return preRun(d)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(d)
		},
	}
	cmd.Flags().StringVar(&d.colorUse, "color", d.colorUse, "Color output (auto/always/never)")
	cmd.Flags().StringVar(&d.output, "output", d.output, "Output decoration (default/gitlab)")
	cmd.Flags().BoolVar(&d.validate, "validate", d.validate,
		"Whether to validate pipeline file against current JSON schema")
	cmd.Flags().StringVar(&d.file, "file", "", "pipeline file to run")
	cmd.Flags().StringArrayVar(&d.vals, "set", d.vals, "set value to data tree prior to run")
	return cmd
}
