# syntax=docker/dockerfile:1

FROM alpine:latest AS builder
RUN mkdir -p /home/nonroot/.config/timpani /home/nonroot/.local/share/timpani && \
    chown -R 65532:65532 /home/nonroot && \
    chmod -R 700 /home/nonroot

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /home/nonroot
USER nonroot:nonroot

COPY --from=builder /home/nonroot/ .
VOLUME ["/home/nonroot/.config/timpani", "/home/nonroot/.local/share/timpani"]

ARG TARGETARCH
COPY --chmod=700 --chown=nonroot:nonroot dist/timpani_linux_${TARGETARCH}*/timpani .
ENTRYPOINT ["./timpani"]
EXPOSE 14480/tcp

HEALTHCHECK --interval=10s --timeout=3s --retries=3 CMD ["./timpani", "--health-check"]
