RMQ Cat
=======

Simple tool for learning Go AMQP library.

Build
-----

Uses [https://github.com/golang/dep](https://github.com/golang/dep) for dependency management.

```
go get -u github.com/golang/dep/cmd/dep
```

*NOTE*: all dependencies committed to the `vendor/` directory.

Include a new dependency
------------------------

after importing into code, run `dep ensure`

Docker
------

Use the included `docker-compose.yml` file to start a RabbitMQ container:

```
docker-compose up -d
```
