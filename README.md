tracking-api-chaos
==================

Exactly like our Tracking API, but with more chaos, and less actually doing stuff.

`tracking-api-chaos` is a fake API server compatible with 
[Segment's Tracking API](https://segment.com/docs/sources/server/http/). It does nothing
but accept requests, validate the input, and write it to a file. Also, you can configure it to fail
chaotically, adding random latency for example, or returning 429s some amount of the time.

Building
========

```sh
make
```

Running
=======

`./tracking-api-chaos -help`


This config will yield 500's to 10% of requests:

```sh
% echo '[{weight: 10, statusCode: {code: 500}}]' |\
  ./tracking-api-chaos -chaos -

...

tracking-api-chaos[69277]: - Causing chaos chaos.StatusCodeChaos{Code:500, Body:[]uint8(nil)}
tracking-api-chaos[69277]: - [::1]:3000->[::1]:51124 - localhost:3000 - POST /v1/batch - 500 Internal Server Error - "analytics-go (version: 3.0.0)"
```

You can also use Docker:

```sh
% docker build -t tracking-api-chaos .

% echo '[{weight: 10, statusCode: {code: 500}}]' | docker run tracking-api-chaos -chaos -
```

Example config
==============

Config is YAML. Each top-level item defines a certain amount of chaos to inject.

The `weight` field is a percentage of requests that will be affected by this item. 
(All other requests will get no chaos.)

The `latency` item causes latency (with some random jitter) by `latency` number of
milliseconds.

The `statusCode` item causes status code of `code` with a body of `body`.

```yaml
- weight: 5
  latency:
    latency: 10000
- weight: 5
  latency:
    latency: 31000
- weight: 5
  statusCode:
    code: 500
    body: "Something went wrong"
- weight: 5
  statusCode:
    code: 429
    body: "slow down"
```
