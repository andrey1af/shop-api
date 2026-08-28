# Supplier E2E tests

The tests exercise the public HTTP API against the real service and PostgreSQL.
They create unique data and remove it at the end of the test.

From the repository root, start the Docker Compose stack and run the suite:

```sh
docker compose up --build --wait
make test-api-service-e2e
```

The API defaults to `http://localhost:8080`. Override it when needed:

```sh
E2E_BASE_URL=http://localhost:9080 make test-api-service-e2e
```

The `e2e` build tag keeps these tests out of the regular `go test ./...` run.
