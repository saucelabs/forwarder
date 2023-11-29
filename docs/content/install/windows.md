---
title: Windows
---

# Install Forwarder on Windows

### Unpack the zip file

{{< tabs "windows-install" >}}
{{< tab "ARM64" >}}
```powershell
mkdir C:\forwarder
Invoke-WebRequest -Uri {{< data "latest" "windows.aarch64.zip" >}} -OutFile forwarder.zip
Expand-Archive -Path forwarder.zip -DestinationPath C:\forwarder
```
{{< /tab >}}
{{< tab "x86-64" >}}
```powershell
mkdir C:\forwarder
Invoke-WebRequest -Uri {{< data "latest" "windows.x86_64.zip" >}} -OutFile forwarder.zip
Expand-Archive -Path forwarder.zip -DestinationPath C:\forwarder
```
{{< /tab >}}
{{< /tabs >}}

### Add the binary to PATH  

Add `C:\forwarder` to `PATH` environment variable

```powershell
$currentPath = [System.Environment]::GetEnvironmentVariable('PATH', [System.EnvironmentVariableTarget]::Machine)
$newPath = "$currentPath;C:\forwarder"
[System.Environment]::SetEnvironmentVariable('PATH', $newPath, [System.EnvironmentVariableTarget]::Machine)
```

### Add completion

Open PowerShell and check if you already have a profile.

```powershell
Test-Path $PROFILE
```

If the command returns `False`, create a new profile.

```powershell
New-Item -ItemType File -Path $PROFILE -Force
```

Add PowerShell completion to the profile.

```powershell
Add-Content -Path $PROFILE -Value ". C:\forwarder\completions\forwarder.ps1"
```

### Edit config file

This step is optional.
You can use default configuration or configure Forwarder with flags or environment variables.
See CLI reference for more details.

```powershell
notepad C:\forwarder\forwarder.yaml
```

### Start Forwarder

```powershell
forwarder.exe run --config-file C:\forwarder\forwarder.yaml
```
