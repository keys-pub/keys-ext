# How to Release

```shell
./scripts/release.sh
```

## Testing

To test the goreleaser configuration:

```shell
cd service
goreleaser --snapshot --skip-publish --rm-dist
```
