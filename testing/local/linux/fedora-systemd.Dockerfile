FROM fedora:latest

RUN dnf -y update
RUN dnf -y install systemd

RUN dnf -y install bash-completion
RUN echo "source /usr/share/bash-completion/bash_completion" >> /root/.bashrc

CMD [ "/sbin/init" ]
