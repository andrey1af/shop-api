FROM postgres:17-alpine

WORKDIR /migrations

COPY migrations/*.sql ./
COPY migrations/run.sh /usr/local/bin/run-migrations

RUN sed -i 's/\r$//' /usr/local/bin/run-migrations /migrations/*.sql \
    && chmod +x /usr/local/bin/run-migrations

ENTRYPOINT ["/usr/local/bin/run-migrations"]
