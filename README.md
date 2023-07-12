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
        "nya": "123"
    },
    "one": {
        "two": {
            "three": {
                "num": 1
            }
        },
        "arr": [
            "a",
            "b"
        ]
    }
}
```

```go
type Out struct {
	One     string      `rjson:"uwu.nya"`           // "123"
	Two     int         `rjson:"one.two.three.num"` // 1
	Three   []string    `rjson:"one.arr"`           // ["a","b"]
    Four    string      `rjson:"one.arr[0]"`        // "a"
}
```