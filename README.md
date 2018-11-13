Light Wallet Service (for FnFn)
==========

Requirements
----------

LWS relies on a few other services to run (all services should be up before LWS which could connect to):

* MySQL Database
* Mqtt
* RabbitMQ
* Redis
* Core wallet

All these services could be configured in `.env` file (sample could be found in `.env.sample`).

Deploy
----------

### Binary

Download binary from [Github release](https://github.com/FissionAndFusion/lws/releases) page.

Or build the binary from repository by yourself:

~~~
GOOS=linux GOARCH=amd64 go build -o "./gateway" cmd/gateway/main.go
GOOS=linux GOARCH=amd64 go build -o "./stream" cmd/stream/main.go
~~~

`GOOS` and `GOARCH` could be specified by your target environment.

Then just start service by execute the binary files with environment variables in `.env` file.

Development
----------

### Package management

Use go module (>= 1.11) to manage packages.

~~~bash
go mod download # download all dependencies
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


### Schema sql
~~~shell
# execute `./test/data/schema.sql`
# if you are using the docker-compose to set up mysql service
mysql --host=127.0.0.1 --port=13307 -u root lws < ./test/data/schema.sql
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
