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
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/kaptinlin/jsonschema"
	xlog "github.com/rkosegi/slog-config"
	ytp "github.com/rkosegi/yaml-pipeline/pkg/pipeline"
	"github.com/rkosegi/yaml-pipeline/pkg/utils"
	"github.com/rkosegi/yaml-pipeline/pkg/version"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/fluent"
	"github.com/rkosegi/yaml-toolkit/props"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
	"gopkg.in/yaml.v3"
)

var pp = props.PathCodec().Parser()

type data struct {
	sc         *xlog.SlogConfig
	file       string
	validate   bool
	pp         ytp.PipelineSpec
	vals       []string
	output     string
	sl         ytp.Listener
	colorUse   string
	outputOpts []string
	start      time.Time
	end        time.Time
	varsFrom   []string
}

var (
	startOpStyle = color.Style{color.FgMagenta}
	endOpStyle   = color.Style{color.FgGreen}
	errStyle     = color.Style{color.FgRed, color.OpBold}
)

func asErr(errMap map[string]*jsonschema.EvaluationError) error {
	// just first error
	for _, v := range errMap {
		return v
	}
	return nil
}

func preRun(d *data) error {
	applyColorUse(d.colorUse)
	if l, ok := ol[d.output]; ok {
		d.sl = l
	} else {
		return fmt.Errorf("no such output decoration: %s, known types : %v", d.output, strings.Join(lo.Keys(ol), ","))
	}
	if len(d.outputOpts) > 0 {
		if si, ok := d.sl.(configurableOutput); ok {
			if err := si.SetOpts(lo.Associate(d.outputOpts, func(opt string) (string, string) {
				parts := strings.SplitN(opt, "=", 2)
				if len(parts) == 1 {
					return parts[0], ""
				}
				return parts[0], parts[1]
			})); err != nil {
				return fmt.Errorf("unable to configure output decoration '%s': %w", d.output, err)
			}
		}
	}
	if d.validate {
		fmt.Fprintln(os.Stderr, color.Blue.Render("[Schema] Validating document"))
		res, err := utils.ValidateFileAgainstSchema(d.file)
		if err != nil {
			return err
		}
		if res != nil && !res.Valid {
			tree := treeprint.New()
			utils.DumpSchemaEvalResultToTree(tree, res.Details)
			fmt.Fprintln(os.Stderr, tree.String())
			return asErr(res.Errors)
		}
		fmt.Fprintln(os.Stderr, color.Blue.Render("[Schema] OK"))
	}
	return nil
}

func loadVarsFromFile(fpath string, gd dom.ContainerBuilder) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	vars := gd.Child("vars").(dom.ContainerBuilder)
	cb, err := dom.DecodeReader(f, fluent.DefaultFileDecoderProvider(fpath))
	if err != nil {
		return err
	}
	cb.AsContainer().Each(func(cn string, cv dom.Node) bool {
		vars.Set(pp.MustParse(cn), cv)
		return false
	})
	return nil
}

func setValues(d *data, gd dom.ContainerBuilder) error {
	gd.AddValue("vars", dom.ContainerNode())
	if len(d.varsFrom) > 0 {
		for _, f := range d.varsFrom {
			fmt.Fprintln(os.Stderr, color.Blue.Sprintf("[Values] Loading %s", f))
			if err := loadVarsFromFile(f, gd); err != nil {
				return err
			}
		}
	}
	fmt.Fprintln(os.Stderr, color.Blue.Render("[Values] Setting values"))
	utils.ApplyVarsToDom(d.pp.Vars, "vars", gd)
	utils.ApplyValues(gd, d.vals)
	fmt.Fprintln(os.Stderr, color.Blue.Render("[Values] OK"))
	return nil
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

func run(ctx context.Context, d *data) error {
	var (
		bytes []byte
		err   error
	)
	d.start = time.Now()
	defer func() {
		d.end = time.Now()
	}()
	if bytes, err = os.ReadFile(d.file); err != nil {
		return err
	}
	if err = yaml.Unmarshal(bytes, &d.pp); err != nil {
		return err
	}
	gd := dom.ContainerNode()
	if err = setValues(d, gd); err != nil {
		return err
	}
	return ytp.New(
		ytp.WithData(gd),
		ytp.WithListener(d.sl),
	).ExecuteContext(ctx, d.pp.ActionSpec)
}

func New() *cobra.Command {
	d := &data{
		validate:   true,
		output:     "default",
		outputOpts: []string{},
		colorUse:   "always",
	}

	short := "Runs a pipeline from a file"
	cmd := &cobra.Command{
		Use:   "yp",
		Short: short,
		Long: short + "\n" + `
File is validated against JSON schema unless validation is explicitly disabled (--validate false).
Initial values can be set using --set keyX=valueY.

Supported output decorations:
.______________________________________________________________________.
| Name    | Options                                                    |
|---------|------------------------------------------------------------+
| default | No options                                                 |
|---------|------------------------------------------------------------+
| github  | No options                                                 |
|---------|------------------------------------------------------------+
| gitlab  | * collapse - true/false - controls section collapse/expand |
|         |------------------------------------------------------------+
|         | * indent-section-title - true/false - controls whether to  |
|         | indent title in section header                             |
+---------+------------------------------------------------------------+

To set output option, use argument like this: --output-opt collapse=true


Initial data can be loaded from set of files using --vars-from:

  --vars-from file1.yaml --vars-from file2.yaml

All files must exists and must be parseable (as a mapping/dictionary).
Data are merged into global data tree in the order they are declared on command line.
Explicit override using --set is processed afterwards, so it takes precedence.
`,
		Version: version.Get(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(d.file) == 0 {
				return errors.New("pipeline file is required (--file)")
			}
			return preRun(d)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), d)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			printStats(d)
		},
	}
	cmd.Flags().StringSliceVar(&d.outputOpts, "output-opt", d.outputOpts, "Output options. These are decoration-specific")
	cmd.Flags().StringVar(&d.colorUse, "color", d.colorUse, "Color output (auto/always/never)")
	cmd.Flags().StringVar(&d.output, "output", d.output, "Output decoration (default/gitlab/github)")
	cmd.Flags().BoolVar(&d.validate, "validate", d.validate,
		"Whether to validate pipeline file against current JSON schema")
	cmd.Flags().StringVar(&d.file, "file", "", "pipeline file to run")
	cmd.Flags().StringArrayVar(&d.vals, "set", d.vals, "set value to data tree prior to run")
	cmd.Flags().StringArrayVar(&d.varsFrom, "vars-from", d.varsFrom, "Additional data from files to load under .vars")
	return cmd
}

func printStats(d *data) {
	fmt.Fprintln(os.Stderr, color.Gray.Sprintf("[Summary] completed in %v", d.end.Sub(d.start)))
}
