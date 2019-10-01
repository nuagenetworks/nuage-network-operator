FROM centos:7

ENV OPERATOR=/usr/local/bin/nuage-network-operator \
    USER_UID=1001 \
    USER_NAME=nuage-network-operator

# install operator binary
COPY build/_output/bin/nuage-network-operator ${OPERATOR}
COPY bindata /bindata
COPY build/bin /usr/local/bin

RUN  /usr/local/bin/user_setup

WORKDIR /
ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
