# simpledns

respond with to queries to with simple responses

to solve some of my more trivial dns needs, i made this simple server, which i
typically use alongside [splitdns](https://github.com/noriah/splitdns).

## usage

```shell
simpledns /path/to/config.json
```

## configuration

example config.json

```json
{
  // listen on localhost at UDP port 8053
  "listen": "127.0.0.1:8053",
  "zones": [
    {
      "name": "my.awesome.tld.",
      "records": [
        {"type": "A", "value": "10.23.34.56"},
        {"type": "AAAA", "value": "ff00::0"},
        {"type": "TXT", "value": "a text value"}
      ]
    },
    {
      "name": "mx.my.awesome.tld.",
      "records": [
        {"type": "MX", "value": "mail.other.tld."},
        {"type": "TXT", "value": "some text"},
        {"type": "TXT", "value": "some more text"}
      ]
    }
    // more if desired
  ]
}
```

*remove comments. [golang json][go-json] follows [strict json spec][rfc7159]*

zone/name order does not matter.

[go-json]: https://pkg.go.dev/encoding/json
[rfc7159]: https://www.ietf.org/rfc/rfc7159.txt
