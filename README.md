# Go Server

A small web server built with only the Go standard library — no
framework, no database, no outside dependencies.

## What it does

- `GET /` — serves the static home page in `static/`
- `GET /hello` — replies with plain text `Hello!`
- `POST /form` — accepts a name and address, stores the submission
  in memory, and shows it back safely escaped
- `GET /submissions` — lists every submission so far

Wrong methods get `405 Method Not Allowed`, and a form with an
empty name or address gets `400 Bad Request`.

Submissions live only in memory: restarting the server clears them.

## Run it

You need Go 1.25 or newer.

```sh
go run .
```

Then open <http://localhost:8080> in your browser.

Use a different port with the `PORT` environment variable:

```sh
PORT=9090 go run .
```

## Tests

```sh
go test ./...
go vet ./...
```

The tests cover `/hello` (GET and wrong method), `/form` (valid
post, wrong method, empty fields, HTML escaping), `/submissions`,
and the `PORT` setting.

## Files

- `main.go` — handlers, the in-memory store, timeouts and graceful
  shutdown
- `main_test.go` — tests using `net/http/httptest`
- `static/` — `index.html`, `form.html` and `style.css`
- `go.mod` — the Go module, no outside dependencies
