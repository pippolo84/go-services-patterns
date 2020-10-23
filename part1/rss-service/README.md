# RSS service

RSS service is a simple RSS aggregator.
You can add a RSS feed to it, list all the added feeds and finally stream:

- Title
- Description
- Content

of each feed, one by one.

### Usage:

1. List all feeds

```
curl --request GET \
  --url http://localhost:8080/feeds

[{"name":"test-feed","url":"http://joeroganexp.joerogan.libsynpro.com/rss"}]
```

2. Add a feed

```
curl --request POST \
  --url http://localhost:8080/feed \
  --header 'content-type: application/json' \
  --data '{
	"name": "test-feed",
	"url": "http://joeroganexp.joerogan.libsynpro.com/rss"
}'
```

3. stream the content of each feed, starting again after one second (N.B.: use `curl` or your browser)

```
curl --request GET \
  --url http://localhost:8080/items

[{"title":"#1552 - Matthew McConaughey","description":"Matthew McConaughey is an Academy Award-winning actor known for such films as Dazed and Confused, The Dallas Buyers Club, Interstellar, Free State of Jones, and the HBO television series True Detective. His new memoir Greenlights is now available everywhere and at https://greenlights.com","content":"Matthew McConaughey is an Academy Award-winning actor known for such films as Dazed and Confused, The Dallas Buyers Club, Interstellar, Free State of Jones, and the HBO television series True Detective. His new memoir Greenlights is now available everywhere and at https://greenlights.com"}...

after 1 second...

[{"title":"#1552 - Matthew McConaughey","description":"Matthew McConaughey is an Academy Award-winning actor known for such films as Dazed and Confused, The Dallas Buyers Club, Interstellar, Free State of Jones, and the HBO television series True Detective. His new memoir Greenlights is now available everywhere and at https://greenlights.com","content":"Matthew McConaughey is an Academy Award-winning actor known for such films as Dazed and Confused, The Dallas Buyers Club, Interstellar, Free State of Jones, and the HBO television series True Detective. His new memoir Greenlights is now available everywhere and at https://greenlights.com"}...

...

```

### Requisites:

- the service should implements suitable timeout to avoid stale connections
- the service should shutdown gracefully when receiving a `SIGINT` (CTRL-C)
- since the `/feed` accept a JSON encoded body, it should check with care what it receives

### What you should do:

1. Complete the missing code in `server.go`. Start with the `/feeds` and `/feed` endpoints.

You can use the unit tests:

`go test -v -race -timeout=30s ./...`

and the integration tests:

`go test -v --tags=integration -race -timeout=120s ./...`

to check the code you write, without trying it manually.

Hint: use `t.Skip()` in front of each test case if you want to temporarily skip it!

2. Complete the missing code in `main.go`.
3. Build and try the service.
4. Add the `/items` endpoint to stream the result. See [this](https://dev.to/mirzaakhena/server-sent-events-sse-server-implementation-with-go-4ck2) if you want an example of a Server Sent Events (SSE) server in Go.
5. Try to stream the content of some feed (look at the comment at the beginning of `main.go` to find some RSS feed to try)
6. Have fun!

### Time to spare?

Have a look ad the `TODO` in `integration_test.go` ;)