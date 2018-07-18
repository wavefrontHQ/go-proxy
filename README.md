# go-proxy [![travis build status](https://travis-ci.com/wavefrontHQ/go-proxy.svg?branch=master)](https://travis-ci.com/wavefrontHQ/go-proxy)
Go Proxy for Wavefront

## To start developing

##### You have a working [Go environment](https://golang.org/doc/install).

```
$ go get -d github.com/wavefronthq/go-proxy
$ cd $GOPATH/src/wavefronthq/go-proxy
$ make
```

## To build packages

#### Linux packages (.deb, .rpm)

##### You have a [Docker environment](https://docs.docker.com/).

```
$ make docker-build 
```

##### You have a Linux installation.

```
$ make package
```

#### All packages (Linux, Windows, Darwin)

```
$ make package-all
```
