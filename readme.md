# GOSTALL

Ever wanted to run `go install ./cmd/server` and not have the binary install in your GOBIN be called `server`?

This is exactly what gostall does

```
go install github.com/davidmdm/gostall

gostall ./cmd/server my-desired-binary-name
```

I won't lie, the implementation is trivial. Perhaps even more simply solved by a bash script.
But this is Go and we like Go, so here we are.
