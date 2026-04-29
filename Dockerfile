FROM alpine:3.21

RUN apk --no-cache add ca-certificates curl && \
    adduser -D -u 1000 ragouser

COPY server /server

USER 1000

EXPOSE 4000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:4000/health || exit 1

ENTRYPOINT ["/server"]
