GIT_ROOT = $(shell git rev-parse --show-toplevel)

dist: clean
	@make -C $(GIT_ROOT) dist

.PHONE: clean
clean:
	@rm -f *.deb *.rpm
	@podman image rm -f debian-systemd:bullseye ||:
	@podman image rm -f fedora-systemd:latest ||:

forwarder.deb:
	@cp $(GIT_ROOT)/dist/forwarder*linux_arm64.deb forwarder.deb

forwarder.rpm:
	@cp $(GIT_ROOT)/dist/forwarder*_linux.aarch64.rpm forwarder.rpm

.PHONY: debian
debian: export IMG = debian-systemd:bullseye
debian: export INSTALL_CMD = apt install -f -y
debian: export EXT = deb
debian: forwarder.deb
	@podman image exists "$(IMG)" || podman build --no-cache -t "$(IMG)" -f "debian-systemd.Containerfile"
	@$(MAKE) .up

.PHONY: fedora
fedora: export IMG = fedora-systemd:latest
fedora: export INSTALL_CMD = dnf install -y
fedora: export EXT = rpm
fedora: forwarder.rpm
	@podman image exists "$(IMG)" || podman build --no-cache -t "$(IMG)" -f "fedora-systemd.Containerfile"
	@$(MAKE) .up

.PHONY: .up
.up:
	@podman run -p 3128:3128 -d -v ./forwarder.$(EXT):/forwarder.$(EXT) --name forwarder-testing --replace $(IMG)
	@podman exec forwarder-testing $(INSTALL_CMD) "/forwarder.$(EXT)"
	@podman cp ./forwarder.yaml forwarder-testing:/etc/forwarder/forwarder.yaml
	@podman exec forwarder-testing systemctl start forwarder

.PHONY: down
down:
	@podman rm --force forwarder-testing

.PHONY: shell
shell:
	@podman exec -it forwarder-testing /bin/bash

.PHONY: status
status:
	@podman exec forwarder-testing systemctl status forwarder

.PHONY: restart
restart:
	@podman exec forwarder-testing systemctl restart forwarder

.PHONY: logs
logs:
	@podman exec -it forwarder-testing journalctl -u forwarder -n all -f
