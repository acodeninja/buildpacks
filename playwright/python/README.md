# acodeninja/playwright-python

A buildpack for installing playwright and caching browser dependencies.

## Usage

If using the `project.toml` file to define buildpacks add the following:

```toml
[[io.buildpacks.group]]
uri = "paketo-buildpacks/python"

[[io.buildpacks.group]]
uri = "acodeninja/playwright-python"
```

If using the command line, add the following:

```shell
pack build --buildpack paketo-buildpacks/python \
           --buildpack acodeninja/playwright-python
```
