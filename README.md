# WIP Learn to use Echo framework

Small example to get familiar with the Echo framework. Coded in 2 hours on a Sunday so quite dirty for now...

## Usage

- build the image `docker build . --no-cache -t auth_server:v1`
- docker-compose up

or

- docker-compose up --build

Test login works `curl --location --request POST 'localhost:1323/api/login' \ --header 'Content-Type: application/json' \ --data-raw '{ "username": "george", "password": "testtest" }'`

## Benchmarks

Quite fast at 16K transactions per second (AMD64 8C/16T | 16GB | SATA SSD)

```console
siege -c512 -r 500 --content-type "application/json" 'http://localhost:1323/api/login POST {"username": "paul", "password": "testtest"}'
** SIEGE 4.0.4
** Preparing 512 concurrent users for battle.
The server is now under siege...
Lifting the server siege...
Transactions:                256000 hits
Availability:                100.00 %
Elapsed time:                 15.93 secs
Data transferred:             94.73 MB
Response time:                 0.03 secs
Transaction rate:          16070.31 trans/sec
Throughput:                    5.95 MB/sec
Concurrency:                 495.42
Successful transactions:     256000
Failed transactions:              0
Longest transaction:           0.53
Shortest transaction:          0.00

```
