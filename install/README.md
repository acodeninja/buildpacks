# ACodeNinja Buildpack for installing system packages

Install packages from the ubuntu apt repository. This buildpack is inspired by the [fagiani/apt](https://registry.buildpacks.io/buildpacks/fagiani/apt) buildpack which is no longer compatible with the latest buildpack API.

## Configuration

Place a file called `.InstallPackages` in the root of the application folder.

```text title=".InstallPackages"
curl
git
postgresql-client
```
