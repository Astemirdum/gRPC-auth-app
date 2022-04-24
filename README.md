# gRPC Service AuthUsersApp на Go

## The following concepts are applied in app:
- Development of a gRPC service based on a proto file in Go:
  - create
  - delete
  - list users
  - authorize user

- Authorization and authentication using a stateless approach (JWT)
- The Clean Architecture approach (dependency injection)
- Postgres database using <a href="https://github.com/jmoiron/sqlx">sqlx</a> library.
- list users, the data is cached in <a href="https://redis.io/">Redis</a> for 1 min.
- Application configuration using <a href="https://github.com/spf13/viper">spf13/viper</a> library.
- Graceful Shutdown.
- Running app in Docker containers.

### to start the service:

#### run app in docker containers
```
make run
```
