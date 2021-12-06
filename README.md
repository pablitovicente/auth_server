# WIP Learn to use Echo framework

Small example to get familiar with the Echo framework. Coded in 2 hours on a Sunday so quite dirty for now...

## Usage

- build the image `docker build . --no-cache -t auth_server:v1`
- docker-compose up

or

- docker-compose up --build

Test login works `curl --location --request POST 'localhost:1323/login' \ --header 'Content-Type: application/json' \ --data-raw '{ "username": "george", "password": "testtest" }'`

## Benchmarks

Quite fast at 17K transactions per second wondering if the benchmarking strategy is wrong...

```console
siege -c512 -t15s --content-type "application/json" 'http://localhost:1323/login POST {"username": "paul", "password": "testtest"}'
** SIEGE 4.0.4
** Preparing 512 concurrent users for battle.
The server is now under siege...
Lifting the server siege...
Transactions:                244092 hits
Availability:                100.00 %
Elapsed time:                 14.10 secs
Data transferred:             9.78 MB
Response time:                0.03 secs
Transaction rate:         17311.49 trans/sec
Throughput:                   0.69 MB/sec
Concurrency:                  508.69
Successful transactions:      244092
Failed transactions:               0
Longest transaction:            0.27
Shortest transaction:           0.00

```
