# Install

Forwarder is a single binary, you can download it from [release page](https://github.com/saucelabs/forwarder/releases).
We provide pre-built packages for x86_64 and arm64 architectures for Linux, macOS and Windows.

## macOS

On macOS you can install Forwarder with [Homebrew](https://brew.sh/):

```bash
brew install saucelabs/tap/forwarder
```

To get the default config file run:

```bash
forwarder run config-file
```

The `config-file` command works with any other command, so you can use it to get the default config file for any command.

Alternatively, you can download the latest `.zip` package from [release page](https://github.com/saucelabs/forwarder/releases).
The binary is signed by Sauce Labs and works out of the box. 

## Linux with Systemd (Debian/Ubuntu/Fedora/CentOS)

Download the latest `.deb` or `.rpm` package from [release page](https://github.com/saucelabs/forwarder/releases) and install it with `dpkg` or `dnf`:

```bash
# Debian/Ubuntu
deb -i forwarder_<version>.linux_<arch>.deb
# Fedora/CentOS
dnf -y install forwarder_<version>.linux_<arch>.rpm
```

Enable and start Forwarder service:

```bash
systemctl enable forwarder
systemctl start forwarder
```

You can configure Forwarder by editing `/etc/forwarder/forwarder.yaml` file.
After editing the config file, restart Forwarder service:

```bash
systemctl restart forwarder
```

## Source

You can also install Forwarder from source:

```go
go get -u github.com/saucelabs/forwarder/cmd/forwarder
```
