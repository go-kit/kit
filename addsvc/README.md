# addsvc

addsvc is an example service, used to illustrate the mechanics of gokit.
It exposes simple functionality on a variety of transports and endpoints.

## Server

To build and run addsvc,

```
$ go install
$ addsvc
```

## Client

addsvc comes with an example client, [addcli][].

[addcli]: https://github.com/go-kit/kit/blob/master/addsvc/client/addcli/main.go

```
$ cd client/addcli
$ go install
$ addcli
```

