package rjson

import (
	"fmt"
	"log"
	"testing"

	assert "github.com/ayes-web/testingassert"
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
	Ten string `rjson:"combined[-].str"`
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

	assert.TestState = t
	assert.Equals(out.One, one)
	assert.Equals(out.Two, two)
	assert.Equals(out.Three, three)
	assert.Equals(out.Four, three[0])
	assert.Equals(out.Five, five)
	assert.Equals(out.Six, []string{"1", "2"})
	assert.Equals(out.Seven, []int{1, 2})
	assert.Equals(out.Eight.Text, one)
	assert.Equals(out.Nine[0].Text, "1")
	assert.Equals(out.Nine[1].Text, "2")
	assert.Equals(out.Ten, "2")
}
