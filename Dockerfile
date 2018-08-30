FROM alpine:latest

RUN apk add bash bind-tools vim tmux curl
CMD /bin/bash
