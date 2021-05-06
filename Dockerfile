ARG RELEASE_IMAGE_NAME=alpine
ARG RELEASE_IMAGE_TAG=3.11.5

FROM ${RELEASE_IMAGE_NAME}:${RELEASE_IMAGE_TAG}

LABEL app="notification-api"
LABEL description="notification API"

RUN apk --no-cache add tzdata ca-certificates

COPY ./serverd /

CMD ./serverd
