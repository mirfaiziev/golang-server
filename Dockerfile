FROM golang:1.21-alpine AS builder

ARG USER=runner

RUN apk add --update --no-cache sudo git make

# add new user
RUN adduser -D $USER && mkdir -p /etc/sudoers.d \
        && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER \
        && chmod 0440 /etc/sudoers.d/$USER

USER $USER

WORKDIR /app



