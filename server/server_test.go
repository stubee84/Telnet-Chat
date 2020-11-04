package server

import (
	"fmt"
	"regexp"
	"testing"
)

func TestRemoveFromSlice(t *testing.T) {
	// // connectedClients = []string{"kallye", "andrew", "stuart"}

	// removeClient("kallye")

	// fmt.Println(connectedClients)
}

func TestRegex(t *testing.T) {
	testStr := `/ignore andrew
`
	// expectedResultSlice := []string{"pm", "andrew", "message to send"}

	if ignore.MatchString(testStr) {
		for i, str := range ignore.FindStringSubmatch(testStr) {
			if i == 1 {
				fmt.Println(str)
			}
		}

	} else {
		t.Fail()
	}

}

func TestSplitString(t *testing.T) {
	testStr := `i
am
a
split
string
`

	strToMatch, _ := regexp.Compile("^i\nam\na\nsplit\nstring\n$")
	if strToMatch.MatchString(testStr) {
		for _, str := range strToMatch.FindStringSubmatch(testStr) {
			fmt.Println(str)
		}
	}
}
