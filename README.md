tracking-api-chaos
==================

Exactly like our Tracking API, but with more chaos, and less actually doing stuff.

`tracking-api-chaos` is a fake API server compatible with 
[Segment's Tracking API](https://segment.com/docs/sources/server/http/). It does nothing
but accept requests, validate the input, and write it to a file. Also, you can configure it to fail
chaotically, adding random latency for example, or returning 429s some amount of the time.

Building
========

```
make
```

Running
=======

`./tracking-api-chaos -help`


This config will yield 500's to 10% of requests:

```
% echo '[{weight: 10, statusCode: {code: 500}}]' |\
  ./tracking-api-chaos -chaos -

...

tracking-api-chaos[69277]: - Causing chaos chaos.StatusCodeChaos{Code:500, Body:[]uint8(nil)}
tracking-api-chaos[69277]: - [::1]:3000->[::1]:51124 - localhost:3000 - POST /v1/batch - 500 Internal Server Error - "analytics-go (version: 3.0.0)"
```
