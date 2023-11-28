---
title: macOS
---

# Install Forwarder on macOS

* [With Homebrew]({{< ref "#with-homebrew" >}})
* [With zip package]({{< ref "#with-zip-package" >}})

## With Homebrew

On macOS you can install Forwarder with [Homebrew](https://brew.sh/):

### Install

```bash
brew tap saucelabs/tap
brew install forwarder
```

### Edit config file

This step is optional.
You can use default configuration or configure Forwarder with flags or environment variables.
See CLI reference for more details.

```bash
forwarder run config-file > forwarder.yaml
vim forwarder.yaml
```

### Start Forwarder

```bash
forwarder run --config-file forwarder.yaml
```

## With zip package

Forwarder provides `.zip` package with a signed binary that can be used on any macOS version.

### Unpack the zip file

```bash
curl -L -o forwarder.zip {{< data "latest" "darwin-signed.all.zip" >}}
sudo mkdir -p /opt/forwarder
sudo unzip -d /opt/forwarder forwarder.zip
```

### Check the signature

Run the following command, you should see `Developer ID Application: SAUCE LABS INC`.

```bash
codesign -dvv /opt/forwarder/forwarder
```

### Link the binary

```bash
sudo ln -s /opt/forwarder/forwarder /usr/local/bin/forwarder
```

### Add completion

{{< tabs "macos-completion" >}}
{{< tab "Zsh" >}}
```bash
echo 'source <(forwarder completion zsh)' >>~/.zshrc
```
{{< /tab >}}
{{< tab "Bash" >}}
```bash
echo 'source <(forwarder completion bash)' >>~/.bash_profile
```
{{< /tab >}}
{{< /tabs >}}

### Edit config file

This step is optional.
You can use default configuration or configure Forwarder with flags or environment variables.
See CLI reference for more details.

```bash
vim /opt/forwarder/forwarder.yaml
```

### Start Forwarder

```bash
forwarder run --config-file /opt/forwarder/forwarder.yaml
```