# Examples

- ### Find latest docker binary for download

[This pipeline](find_latest_docker_binary.yaml) checks https://download.docker.com/linux/static/stable/x86_64/ and find latest binary for download.

- ### Find latest Helm chart version from Helm repository

[This pipeline](find_latest_chart_version.yaml) fetches index.yaml from given Helm repository and
find latest version of chart that matches provided regular expression.

- ### Rename files in directory based on file content

[This pipeline](rename_gpx_file_date_first_elem.yaml) list all GPX files in given directory and rename them according to `time` element within the first `trkpt` element.

- ### Merge JSON schema store catalogs

[This pipeline](https://github.com/rkosegi/json-schemas/blob/main/pipelines/render.yaml) merges public
[JSON schema store catalog](https://www.schemastore.org/api/json/catalog.json) with local one.

- ### Copy file from GitHub release into other git repo that hosts JSON schema store catalog

[This pipeline](https://github.com/rkosegi/json-schemas/blob/main/pipelines/embed.yaml) fetches file from existing GitHub release assets
and puts it into local directory. Then it updates JSON schema store catalog with this new entry.
