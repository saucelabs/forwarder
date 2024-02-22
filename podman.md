# Using Podman instead of Docker

You can use Podman instead of Docker as a container engine.
Docker Compose needs to be installed as we don't support Podman Compose yet.

## Configuration

Make sure you have installed and started Podman and Docker Compose, then follow these steps:

1. Link `podman` command as `docker` command:
  ```
  ln -s $(which podman) /usr/local/bin/docker
  ```
1. Edit `~/.docker/config.json` by changing the `credsStore` to `credStore`, it should look like this:
  ```
  {
    "credStore": "desktop"
  }
  ```
1 .If not on Linux, ssh into the Podman Machine:
  ```
  podman machine ssh
  ``` 
1. Edit the `delegate.conf` file:
  ```
  sudo vi /etc/systemd/system/user@.service.d/delegate.conf 
  ```
  By adding `cpuset` to the `Delegate` line, it should look like this:
  ```
  [Service]
  Delegate=memory pids cpu cpuset io
  ```
1. Restart the Podman Machine
    ```
    podman machine restart
    ```
