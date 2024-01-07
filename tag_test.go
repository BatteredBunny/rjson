package rjson

import (
	"fmt"
	"os"
	"testing"

	assert "github.com/BatteredBunny/testingassert"
)

type testStruct struct {
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
	Ten      string     `rjson:"combined[-].str"`
	Eleven   string     `rjson:"nestedarr[0].test[0].test"`
	Elevenv2 string     `rjson:"nestedarr[-].test[-].test"`
	Twelve   [][]string `rjson:"nesteditter[].thing[].a"`
	Twelvev2 [][]string `rjson:"nesteditterv2[].thing.thingv2.thingv3[].a.value"`
	Thirteen []string   `rjson:"badges[].metadata.value"`
	Fourteen struct {
		Eight struct {
			Text string `rjson:"nya"`
		} `rjson:"uwu"`
	} `rjson:"."`
}

func TestTag(t *testing.T) {
	one := ">_<"
	two := 1
	three := []string{"mrow", "OWO"}
	five := "asdasdasdasd"

	bs, err := os.ReadFile("test.json")
	if err != nil {
		t.Fatal(err)
	}

	rawJson := fmt.Sprintf(string(bs), one, five, two, three[0], three[1])

	var out testStruct
	if err := Unmarshal([]byte(rawJson), &out); err != nil {
		t.Fatal(err)
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
	assert.Equals(out.Eleven, "aaa")
	assert.Equals(out.Elevenv2, "aaa")
	assert.Equals(out.Twelve, [][]string{{"1", "3"}, {"1", "4"}})
	assert.Equals(out.Twelvev2, [][]string{{"1", "3"}, {"1", "4"}})
	assert.Equals(out.Thirteen, []string{"Verified"})
	assert.Equals(out.Fourteen.Eight.Text, one)
}
