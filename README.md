# gRPC Service AuthUsersApp на Go

## В работе применены следующие концепции:
- Разработка gRPC сервиса на основе proto файла на Go: добавление, удаление, авторизация, аутентификация пользователя, список пользователей.
- Авторизация и аутентификация используя stateless подход. Работа с JWT. Применение AuthInterceptor при удаление пользователя.
- Подход Чистой Архитектуры в построении структуры приложения. Dependency injection.
- Работа с БД Postgres, используя библиотеку <a href="https://github.com/jmoiron/sqlx">sqlx</a>.
- На запрос получения списка пользователей данные кешируются в <a href="https://redis.io/">Redis</a> на 1 мин.
- При добавлении пользователя добовляется сообщение в очередь <a href="https://kafka.apache.org/">Kafka</a>.
- Конфигурация приложения с помощью библиотеки <a href="https://github.com/spf13/viper">spf13/viper</a>.
- Graceful Shutdown.
- Запуск из Docker.

### Для запуска сервиса:

#### from Docker
```
make run
```

Если приложение запускается впервые, необходимо применить миграции к базе данных:
```
make migrate
```

#### gRPC Server
```
go run server/cmd/main.go
```

#### gRPC Client
```
go run client/main.go
```

#### Redis
```
src/redis-server
```

#### Kafka
```
bin/zookeeper-server-start.sh config/zookeeper.properties
bin/kafka-server-start.sh config/server.properties
```