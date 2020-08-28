# How to Release

- Build and test with `./scripts/test/all.sh`.
- Create branch with version v1.2.3, the github action will build the apps.

## Test goreleaser

To test the goreleaser configuration:

```shell
cd service
goreleaser --snapshot --skip-publish --rm-dist
```
