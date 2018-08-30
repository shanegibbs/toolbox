FROM alpine:latest

RUN apk add bash bind-tools vim tmux
CMD /bin/bash
