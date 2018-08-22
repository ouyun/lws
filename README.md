Light Wallet Service (for FnFn)
==========

Development
----------

### Package management

Since Go hasn't a best package management solution until v1.11 (2018-08-31), we use [**govendor**](https://github.com/kardianos/govendor) for managing packages simply and temporarily.

~~~bash
go get -u github.com/kardianos/govendor # install govendor tool

govendor sync # sync all packages defined in vendor/vendor.json from remote repositories

govendor fetch github.com/user/xxx-repo # something like `go get` but fetch package into vendor/ and update vendor.json file
~~~
