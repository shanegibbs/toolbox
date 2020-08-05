ARG IMAGE
FROM sham AS sham

FROM ${IMAGE}
ARG IMAGE
ARG USER_ID
ARG SHAM_INIT_OPTIONS

COPY --from=sham /sham/setup /sham-setup
COPY --from=sham /sham/init /sham-init
COPY --from=sham /sham/run /sham-run

RUN printenv

#RUN echo abc $SHAM_INIT_OPTIONS 123 |tee /tmp/options.json
#RUN cat /tmp/options.json
RUN /sham-setup

LABEL com.gibbsdevops.sham.image="${IMAGE}"

USER ${USER_ID}
ENTRYPOINT ["/sham-init"]
CMD ["/bin/bash"]
