# CDN
[![CircleCI](https://circleci.com/gh/bign8/cdn.svg?style=svg)](https://circleci.com/gh/bign8/cdn)
[![Go Report Card](https://goreportcard.com/badge/github.com/bign8/cdn)](https://goreportcard.com/report/github.com/bign8/cdn)

This project is to demonstrate the internals of a CDN.

Topics are based on [Algorithmic Nuggets in Content Delivery](https://people.cs.umass.edu/~ramesh/Site/HOME_files/CCRpaper_1.pdf)

## Setup

1. Download, Install, and Configure the following tools
  * [Go](https://golang.org/dl/)
  * [Docker](https://www.docker.com/products/overview)
  * [Docker Compose](https://docs.docker.com/compose/install/)
2. Run `./run.sh` to build the source, generate containers, and start the environment

## Pieces

* [Origin](origin/README.md) is a website generator
* [Server](server/README.md) is a CDN server
* [Client](client/README.md) is a website consumer
