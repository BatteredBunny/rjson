package main

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/ayes-web/testingassert"
)

type Out struct {
	One   string   `rjson:"uwu.nya"`
	Two   int      `rjson:"one.two.three.num"`
	Three []string `rjson:"one.arr"`
	Four  any      `rjson:"one.two.three"`
}

type RecursiveJson struct {
	Uwu struct {
		Nya string `json:"nya"`
	} `json:"uwu"`
	One struct {
		Two struct {
			Three struct {
				Num int `json:"num"`
			} `json:"three"`
		} `json:"two"`
		Arr []string `json:"arr"`
	} `json:"one"`
}

func TestTag(t *testing.T) {
	one := ">_<"
	two := 1
	three := []string{"mrow", "OWO"}

	j := RecursiveJson{}
	j.Uwu.Nya = one
	j.One.Two.Three.Num = two
	j.One.Arr = three

	bs, err := json.Marshal(j)
	if err != nil {
		log.Fatal(err)
	}

	var out Out
	if err = Unmarshal(bs, &out); err != nil {
		log.Fatal(err)
	}

	testingassert.AssertEquals(t, out.One, one)
	testingassert.AssertEquals(t, out.Two, two)
	testingassert.AssertEqualsDeep(t, out.Three, three)
	t.Log(out.Four)
}
