# Redsync

[![Build Status](https://travis-ci.org/vyng/redsync.svg?branch=master)](https://travis-ci.org/vyng/redsync)

Redsync provides a Redis-based distributed mutual exclusion lock implementation for Go as described in [this post](http://redis.io/topics/distlock).

## Prerequisites
* [Redigo](https://github.com/gomodule/redigo) (Looking for help with removing this)

## Installation

Install Redsync using the go get command:

    $ go get github.com/vyng/redsync

## Documentation

- [Reference](https://godoc.org/gopkg.in/redsync.v1)

## Contributing

Contributions are welcome.

## License

Redsync is available under the [BSD (3-Clause) License](https://opensource.org/licenses/BSD-3-Clause).

## Disclaimer

This code implements an algorithm which is currently a proposal, it was not formally analyzed. Make sure to understand how it works before using it in production environments.
