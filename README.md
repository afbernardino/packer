# Packer

Computes the number of packs that need to be shipped to the customer given the pack sizes and the order size.

## Context

Imagine for a moment that one product line ships in various pack sizes:

* 250 Items
* 500 Items
* 1000 Items
* 2000 Items
* 5000 Items

Customers can order any number of these items through the website, but they will always only be given complete packs.

1. Only whole packs can be sent. Packs cannot be broken open.
2. Within the constraints of Rule 1 above, send out no more items than necessary to fulfil the order.
3. Within the constraints of Rules 1 & 2 above, send out as few packs as possible to fulfil each order.

## Example

| Items ordered | Correct number of packs | Incorrect number of packs            |
|---------------|-------------------------|--------------------------------------|
| 1             | 1 x 250                 | 1 x 500 – more items than necessary  |
| 250           | 1 x 250                 | 1 x 500 – more items than necessary  |
| 501           | 1 x 500                 | 1 x 1000 – more items than necessary |
|               | 1 x 250                 | 3 x 250 – more packs than necessary  |
| 12001         | 2 x 5000                | 3 x 250 – more packs than necessary  |
|               | 1 x 2000                |                                      |
|               | 1 x 250                 |                                      |

## API Specification

See [here](openapi.yaml).

## Prerequisites

* [Go](https://go.dev/) as the default language
* [Docker](https://www.docker.com/) to run the project

## Test

To run the tests run the following command (from the root directory):

```shell
make test
```

### Integration tests

To run the integration tests locally make sure that you have the service running.

You can do that by running the following command (from the root directory):

```shell
docker compose -f docker-compose.yml --env-file ./.env.test up -d --build
```

Then, run the integration tests, by running the following command (from the root directory):

```shell
make integration_test
```

Finally, you can stop the service by running the following command (from the root directory):

```shell
docker compose -f docker-compose.yml --env-file ./.env.test down
```

## Build

To build the service run the following command (from the root directory):

```shell
make build
```