FROM debian:bullseye

RUN apt-get update && apt-get install -y systemd systemd-sysv && apt-get clean
RUN systemctl mask systemd-logind systemd-udevd

CMD ["/lib/systemd/systemd"]
