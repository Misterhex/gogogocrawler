package crawlers

import (
	"strings"
	"testing"
)

func TestShouldReturnTrueIfPrefixIsMp4OrFlv(t *testing.T) {

	var s = "http://play44.net/embed.php?w=600&h=438&vid=Z/zombie-loan-01.flv"

	var result = strings.HasSuffix(s, "flv")

	if !result {
		t.Error()
	}

}

package example

import (
	"fmt"
	"testing"
)



func TestShouldReturnDistinctValues(c) {

	var c := make([]chan int)

	c <- 1
	c <- 1
	c <- 2
	c <- 2
	c <- 3

	set := make(map[int]int)


	for e := range c {

		if _, ok := set[1]; ok {
			
		}
	}
}
