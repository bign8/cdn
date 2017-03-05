# CDN
This project is to demonstrate the internals of a CDN.

Topics are based on [Algorithmic Nuggets in Content Delivery](https://people.cs.umass.edu/~ramesh/Site/HOME_files/CCRpaper_1.pdf)

## Setup

1. Download, Install, and Configure the following tools
  * [Go](https://golang.org/dl/)
  * [Docker](https://www.docker.com/products/overview)
  * [Docker Compose](https://docs.docker.com/compose/install/)
2. Run `make` to generate containers
3. Run `docker-compose up` to start the environment
4. Run `docker-compose scale origin=5` to scale nodes as desired

## Pieces

* [Origin](origin/README.md) is a website generator
* [Server](server/README.md) is a CDN server
* [Client](client/README.md) is a website consumer
