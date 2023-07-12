# rjson

a jq like library for golang that helps simplify recursive json

```
go get github.com/ayes-web/rjson
```


## Parse json such as this without any anonymous structs
For a full example have a look at `tag_test.go`

```json
{
    "uwu": {
        "nya": "\u003e_\u003c"
    },
    "one": {
        "two": {
            "three": {
                "num": 1
            }
        },
        "arr": [
            "mrow",
            "OWO"
        ]
    }
}
```

```go
type Out struct {
	One   string   `rjson:"uwu.nya"`
	Two   int      `rjson:"one.two.three.num"`
	Three []string `rjson:"one.arr"`
}
```