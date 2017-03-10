FROM scratch
ENTRYPOINT ["/app/bin/rinq-httpd"]

EXPOSE 8080

ENV GODEBUG           netdns=cgo
ENV RINQ_AMQP_DSN     amqp://broker
ENV RINQ_HTTPD_BIND   :8080
ENV RINQ_HTTPD_ORIGIN *
ENV RINQ_HTTPD_PING   10

COPY artifacts/build/release/linux/amd64/rinq-httpd /app/bin/
