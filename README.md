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

### Environment Variables

We are using [godotenv](https://github.com/joho/godotenv) to manange variables in `.env` file. Notice that to use it correctly, you should specify relative path in load function:

~~~go
godotenv.Overload("../../.env")
~~~

Otherwise it will fail when your entry file is not under package root directory.

### Database

#### Migration

Up:

~~~
go run cmd/db/migrate.go
~~~

Down:

~~~
go run cmd/db/migrate.go -1
~~~
