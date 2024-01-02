---
title: Linux
---

# Install Forwarder on Linux

* [Debian/Ubuntu]({{< ref "#debianubuntu" >}})
* [RedHat/CentOS/Fedora]({{< ref "#redhatcentosfedora" >}})
* [Generic]({{< ref "#generic" >}})

## Debian/Ubuntu

Forwarder provides `.deb` package with Systemd service for [forwarder run](cli/forwarder_run.md) command.
Other commands are available as well, but you will need to start them manually.

### Install package

{{< tabs "debian-install" >}}
{{< tab "ARM64" >}}
```bash
curl -L -o forwarder.deb {{< data "latest" "linux_arm64.deb" >}}
sudo dpkg -i forwarder.deb
```
{{< /tab >}}
{{< tab "x86-64" >}}
```bash
curl -L -o forwarder.deb {{< data "latest" "linux_amd64.deb" >}}
sudo dpkg -i forwarder.deb
```
{{< /tab >}}
{{< /tabs >}}

### Edit config file

```bash
sudo vim /etc/forwarder/forwarder.yaml
```

### Enable and start Forwarder service

```bash
sudo systemctl enable forwarder
sudo systemctl start forwarder
```

### Check Forwarder status

```bash
sudo systemctl status forwarder
```

## RedHat/CentOS/Fedora

Forwarder provides `.rpm` package with Systemd service for [forwarder run](cli/forwarder_run.md) command.
Other commands are available as well, but you will need to start them manually.

### Install package

{{< tabs "redhat-install" >}}
{{< tab "ARM64" >}}
```bash
sudo rpm -i {{< data "latest" "linux.aarch64.rpm" >}}
```
{{< /tab >}}
{{< tab "x86-64" >}}
```bash
sudo rpm -i {{< data "latest" "linux.x86_64.rpm" >}}
```
{{< /tab >}}
{{< /tabs >}}

### Edit config file

```bash
sudo vim /etc/forwarder/forwarder.yaml
```

### Enable and start Forwarder service

```bash
sudo systemctl enable forwarder
sudo systemctl start forwarder
```

### Check Forwarder status

```bash
sudo systemctl status forwarder
```

## Generic

Forwarder provides `.tar.gz` package with a statically linked binary that can be used on any modern Linux distribution.

### Unpack the tarball

{{< tabs "linux-install" >}}
{{< tab "ARM64" >}}
```bash
curl -L -o forwarder.tar.gz {{< data "latest" "linux.aarch64.tar.gz" >}}
sudo mkdir -p /opt/forwarder
sudo tar -C /opt/forwarder -xzf forwarder.tar.gz
```
{{< /tab >}}
{{< tab "x86-64" >}}
```bash
curl -L -o forwarder.tar.gz {{< data "latest" "linux.x86_64.tar.gz" >}}
sudo mkdir -p /opt/forwarder
sudo tar -C /opt/forwarder -xzf forwarder.tar.gz
```
{{< /tab >}}
{{< /tabs >}}

### Link the binary

```bash
sudo ln -s /opt/forwarder/forwarder /usr/local/bin/forwarder
```

### Add bash completion

{{< tabs "linux-completion" >}}
{{< tab "User" >}}
```bash
echo 'source <(forwarder completion bash)' >>~/.bash_profile
```
{{< /tab >}}
{{< tab "System" >}}
```bash
sudo mkdir -p /etc/bash_completion.d
sudo ln -s /opt/forwarder/completions/forwarder.bash /etc/bash_completion.d/forwarder
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
