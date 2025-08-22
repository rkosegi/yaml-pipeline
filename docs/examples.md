# Examples

- ### Find latest docker binary for download

[This pipeline](find_latest_docker_binary.yaml) checks https://download.docker.com/linux/static/stable/x86_64/ and find latest binary for download.

- ### Find latest Helm chart version from Helm repository

[This pipeline](find_latest_chart_version.yaml) fetches index.yaml from given Helm repository and
find latest version of chart that matches provided regular expression.

- ### Rename files in directory based on file content

[This pipeline](rename_gpx_file_date_first_elem.yaml) list all GPX files in given directory and rename them according to `time` element within the first `trkpt` element.
