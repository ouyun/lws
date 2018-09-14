Light Wallet Service (for FnFn)
==========

Development
----------

### Package management

Since Go hasn't a best package management solution until v1.11 (2018-08-31), we use [**dep**](https://golang.github.io/dep) (officially supported) for managing packages simply and temporarily.

~~~bash
brew install dep # install dep tool

dep init # migrate depedencies from `vendor/` by other tools

dep ensure # sync and install all dependencies into `vender/`

dep ensure -add github.com/user/xxx-repo # something like `go get` but fetch package into `vendor/` and update Gopkg.toml/Gopkg.lock file
~~~

### Related 3rd-party Services

LWS requires MySQL, mqtt(mock server), RabbitMQ and redis as related services. In local development, we use Docker to manage them:

~~~shell
# where docker-compose.yml located
cd test/mock
# start services
docker-compose up -d
# stop services
docker-compose down
~~~

### Environment Variables

We are using [godotenv](https://github.com/joho/godotenv) to manange variables in `.env` file (no required). To install the command-line tool:

~~~shell
go get github.com/joho/godotenv/cmd/godotenv
~~~

Then in local development:

~~~shell
# run
godotenv go run path/to/main.go
# test
godotenv go test ./...
~~~

In CI test, the environment variables should be configured in the CI panel.

In production, the environment variables should be used as normal env way or configured in CD panel.

### Database

#### Migration

Up:

~~~shell
go run cmd/db/migrate.go
~~~

Down:

~~~shell
go run cmd/db/migrate.go -1
~~~
