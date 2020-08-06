ARG IMAGE
ARG SHAM_BINARY_IMAGE
FROM ${SHAM_BINARY_IMAGE} AS sham

FROM ${IMAGE}
ARG IMAGE
ARG USER_ID
ARG SHAM_INIT_OPTIONS

COPY --from=sham /sham/setup /sham-setup
COPY --from=sham /sham/init /sham-init
COPY --from=sham /sham/run /sham-run

RUN /sham-setup

USER ${USER_ID}
ENTRYPOINT ["/sham-init"]
CMD []
