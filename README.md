# **Simple HTTP Multiplexer**

**Functional reqs:**

- an HTTP-server with one handler func;
- HTTP server has limit for 100 inbound requests, single request processing time = 10s;
- HTTP server handles only POST requests with url list (JSON format), urls count < 20;
- HTTP server requests each url and returns results to client as JSON;
- url fetch workers count <= 4, fetch timeout = 1s;
- client cancellation for request processing;
- graceful shutdown

**Nonfunctional reqs:**

- go version > 1.13
- only standart go packages
- deploy and start with docker-compose

To start:

`git clone https://github.com/quantum0cat/simple-http-mux.git`

`cd simple-http-mux`

`docker-compose up`

To stop:

`docker-compose stop simple-http-mux`


**Example curl request:**

`curl -X POST http://localhost:10000 -H 'Content-Type: application/json' \
-d '{
    "urls":[
        "http://ya.ru",
        "http://yandex.ru",
        "http://google.com",
        "http://vk.com",
        "http://twitter.com",
        "http://instagram.com"
    ]
}'`