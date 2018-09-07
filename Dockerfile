FROM alpine:latest

ENV PS1="[toolbox:\w]$ "
RUN apk add bash bind-tools vim tmux curl
CMD /bin/bash
