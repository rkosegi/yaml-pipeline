# YAML pipeline

[![codecov](https://codecov.io/gh/rkosegi/yaml-pipeline/graph/badge.svg?token=BG1D2QKXRE)](https://codecov.io/gh/rkosegi/yaml-pipeline)
[![Go Report Card](https://goreportcard.com/badge/github.com/rkosegi/yaml-pipeline)](https://goreportcard.com/report/github.com/rkosegi/yaml-pipeline)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-pipeline&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-pipeline)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-pipeline&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-pipeline)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-pipeline&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-pipeline)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-pipeline&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-pipeline)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-pipeline&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-pipeline)
[![Go Reference](https://pkg.go.dev/badge/github.com/rkosegi/yaml-pipeline.svg)](https://pkg.go.dev/github.com/rkosegi/yaml-pipeline)
[![Apache 2.0 License](https://badgen.net/static/license/Apache2.0/blue)](https://github.com/rkosegi/yaml-pipeline/blob/main/LICENSE)
[![CodeQL Status](https://github.com/rkosegi/yaml-pipeline/actions/workflows/codeql.yaml/badge.svg)](https://github.com/rkosegi/yaml-pipeline/security/code-scanning)
[![CI Status](https://github.com/rkosegi/yaml-pipeline/actions/workflows/ci.yaml/badge.svg)](https://github.com/rkosegi/yaml-pipeline/actions/workflows/ci.yaml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/rkosegi/yaml-pipeline/badge)](https://scorecard.dev/viewer/?uri=github.com/rkosegi/yaml-pipeline)


It's like scripting, but described in YAML.

Check [examples](docs/examples.md) for more info.

## How to use it

1. declare schema for auto-completion in your editor

    ```yaml
    # yaml-language-server: $schema=https://raw.githubusercontent.com/rkosegi/yaml-pipeline/refs/heads/main/schemas/pipeline.json
    ---
    ```

2. declare input variables (optional)
    ```yaml
    vars:
      mymsg: Hello world!
    ```

3. define steps
    ```yaml
   steps:
     print-vars:
       log:
         message: 'Message is: {{ .vars.mymsg }}'
    ```

4. run pipeline
    ```shell
   yp --file mypipeline.yaml
    ```
