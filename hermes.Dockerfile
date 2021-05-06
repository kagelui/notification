ARG BUILDER_IMAGE_NAME=golang
ARG BUILDER_IMAGE_TAG=1.15-alpine
ARG RELEASE_IMAGE_NAME=alpine
ARG RELEASE_IMAGE_TAG=3.9

FROM ${BUILDER_IMAGE_NAME}:${BUILDER_IMAGE_TAG} as builder

ARG APP_HOME=/app

RUN apk add --update make ca-certificates libc6-compat

WORKDIR $APP_HOME
COPY . $APP_HOME

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go clean -mod=vendor -i -x -cache ./... \
	&& go build -mod=vendor -v ./cmd/hermes

FROM ${RELEASE_IMAGE_NAME}:${RELEASE_IMAGE_TAG}

LABEL app="notification-hermes"
LABEL description="retry callback job"

COPY --from=builder /app/hermes /root
COPY crontab /crontab
COPY script.sh /script.sh
COPY entry.sh /entry.sh
RUN touch /var/log/script.log
RUN chmod 755 /script.sh /entry.sh /root/hermes /var/log/script.log
RUN /usr/bin/crontab /crontab

CMD ["/entry.sh"]
