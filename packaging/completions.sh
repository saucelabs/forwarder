#!/bin/sh

set -e

rm -rf completions
mkdir completions

for sh in bash zsh powershell fish; do
  go run ./cmd/forwarder completion "$sh" >"completions/forwarder.$sh"
done

# Set powershell extension to ps1.
mv completions/forwarder.powershell completions/forwarder.ps1