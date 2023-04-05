FROM --platform=$TARGETPLATFORM scratch
COPY --chown=0:0 bin/webdav-server .tmp/docker-dirs/. /
USER 1000:1000
ENV CONTENT_ROOT /content
ENV LISTEN_ADDR :8080
ENV USERS_FILE /conf/users.yaml
EXPOSE 8080
ENTRYPOINT [ "/webdav-server" ]
