<h1 align="center">rjson</h1> 

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

## rjson syntax explanation

### Path seperator: "."
- Dot symbol is used as path seperator, e.g `one.two.three`

### Array index: [0]
- You can index slices/array like you would normally, e.g `arr[0]`, `arr[1]`

### Last value: [-]
- You can access the last value of an slices/array using this, e.g `arr[-]`

### Value iterator: []

- e.g `arr[].text`

```json
[
    {
        "text": "1",
        "num": 1
    },
        {
        "text": "2",
        "num": 2
    }
]
```
```json
[
    {
        "text": "1",
    },
    {
        "text": "2",
    }
]
```