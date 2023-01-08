FROM quay.io/rockylinux/rockylinux:9

ENV LOG_LEVEL=info

RUN dnf update -y && dnf install -y fdo-admin-cli && dnf clean all

RUN mkdir -p /etc/fdo
RUN chmod -R 775 /etc/fdo

ENTRYPOINT ["/usr/libexec/fdo/fdo-admin-tool"]