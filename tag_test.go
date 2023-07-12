package rjson

import (
	"fmt"
	"log"
	"testing"

	"github.com/ayes-web/testingassert"
)

type outstruct struct {
	One   string   `rjson:"uwu.nya"`
	Two   int      `rjson:"one.two.three.num"`
	Three []string `rjson:"one.arr"`
	Four  string   `rjson:"one.arr[0]"`
	Five  string   `rjson:"jarray[0].mrow"`
	Six   []string `rjson:"combined[].str"`
	Seven []int    `rjson:"combined[]num"`
	Eight struct {
		Text string `rjson:"nya"`
	} `rjson:"uwu"`
	Nine []struct {
		Text string `rjson:"str"`
	} `rjson:"combined"`
}

func TestTag(t *testing.T) {
	one := ">_<"
	two := 1
	three := []string{"mrow", "OWO"}
	five := "asdasdasdasd"

	bs := []byte(fmt.Sprintf(`
	{
		"uwu": {
			"nya": "%s"
		},
		"combined": [
			{
				"str": "1",
				"num": 1,
				"r": {
					"different": "mrp"
				}
			},
			{
				"str": "2",
				"num": 2,
				"r": {
					"fields": "mrp"
				}
			}
		],
		"jarray": [
			{
				"mrow": "%s"
			},
			{
				"1": "uwu"
			}
		],
		"one": {
			"two": {
				"three": {
					"num": %d
				}
			},
			"arr": ["%s", "%s"]
		}
	}
	`, one, five, two, three[0], three[1]))

	var out outstruct
	if err := Unmarshal(bs, &out); err != nil {
		log.Fatal(err)
	}

	testingassert.AssertEquals(t, out.One, one)
	testingassert.AssertEquals(t, out.Two, two)
	testingassert.AssertEqualsDeep(t, out.Three, three)
	testingassert.AssertEquals(t, out.Four, three[0])
	testingassert.AssertEquals(t, out.Five, five)
	testingassert.AssertEqualsDeep(t, out.Six, []string{"1", "2"})
	testingassert.AssertEqualsDeep(t, out.Seven, []int{1, 2})
	testingassert.AssertEquals(t, out.Eight.Text, one)
	testingassert.AssertEquals(t, out.Nine[0].Text, "1")
	testingassert.AssertEquals(t, out.Nine[1].Text, "2")
}
