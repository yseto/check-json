# check-json

``` sh
$ go run ./main.go -query 'select(.level | contains("fatal"))' -log-file testdata/test.log  -output '.timestamp+" "+.msg+" "+.caller' -no-state
"2023-07-27 12:31:55.185 Z Failed to apply database migrations. sqlstore/store.go:173"
```

---

referred to https://github.com/mackerelio/go-check-plugins/blob/master/check-log/lib/check-log.go
