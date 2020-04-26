# fido2

gRPC Service for go-libfido2.
Uses github.com/keys-pub/go-libfido2.

## Dependencies

### Linux

```shell
sudo apt install software-properties-common
sudo apt-add-repository ppa:yubico/stable
sudo apt update
sudo apt install libfido2-dev
```

### macOS

```shell
brew install libfido2
```

### Windows

```shell
scoop bucket add keys.pub https://github.com/keys-pub/scoop-bucket
scoop install libfido2
```
