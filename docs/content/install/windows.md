---
title: Windows
---

# Install Forwarder on Windows

### Install Forwarder

Run the following command to install Forwarder:

```powershell
winget install SauceLabs.forwarder
```

Open a new PowerShell window after installation to use the `forwarder` command.

### Add PowerShell command completion

Run the following script to add command completion to PowerShell:

```powershell
if (-Not (Test-Path -Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force
}
Add-Content -Path $PROFILE -Value "Invoke-Expression (forwarder completion powershell | Out-String)"
```

Open a new PowerShell window to use the `forwarder` command with completion.

### Edit config file

This step is optional.
You can use default configuration or configure Forwarder with flags or environment variables.
See CLI reference for more details.

```powershell
forwarder run config-file > forwarder.yaml
notepad forwarder.yaml
```

### Start Forwarder

```powershell
forwarder run --config-file C:\forwarder\forwarder.yaml
```
