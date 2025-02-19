# ACodeNinja Buildpack for running shell scripts

**Buildpack id**: `acodeninja/run`

This buildpack allows for running custom commands as part of the build process for Cloud Native Buildpacks.

This buildpack is inspired by the [fagiani/run](https://registry.buildpacks.io/buildpacks/fagiani/heroku-buildpack-run) buildpack but with some usability enhancements.

The buildpack must be placed after buildpacks that install the tools required by your script. When running the provided script the buildpack will load all environment variables created by previous buildpacks 

## Configuration

Place a bash script called `buildpack.run.sh` in the root of the application folder.

```shell title="buildpack.run.sh"
#!/usr/bin/env bash

python manage.py collectstatic
```

Add the buildpack `acodeninja/run` to your list of buildpacks.

```toml title="project.toml"
[[io.buildpacks.group]]
uri = "acodeninja/run"
```

### Customisation

| Variable                 | Default            | Description                                                        |
|--------------------------|--------------------|--------------------------------------------------------------------|
| `BP_RUN_SCRIPT_LOCATION` | `buildpack.run.sh` | The path of the shell script relative to the root of the codebase. |
