# syntax=docker/dockerfile:1

FROM alpine:latest AS builder
RUN mkdir -p /home/nonroot/.config /home/nonroot/.local/share

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /home/nonroot
USER nonroot:nonroot

COPY --from=builder --chmod=700 --chown=nonroot:nonroot /home/nonroot/ .
VOLUME ["/home/nonroot/.config", "/home/nonroot/.local/share"]

ARG TARGETARCH
COPY --chmod=700 --chown=nonroot:nonroot dist/timpani_linux_${TARGETARCH}*/timpani .
ENTRYPOINT ["./timpani"]
EXPOSE 14480/tcp

HEALTHCHECK --interval=30s --timeout=3s --retries=2 CMD ["./timpani", "--health-check"]
