package ioreplacer_test

import (
	"fmt"
	"github.com/suin/ioreplacer"
	"io/ioutil"
	"os"
)

func ExampleReader() {
	file, _ := os.Open("./example_test_1.txt")

	replacer := ioreplacer.NewReader(file, map[string]string{
		"apple":     "pumpkin",
		"pineapple": "carrot",
	})

	contents, _ := ioutil.ReadAll(replacer)

	fmt.Printf("%s", contents)

	// OUTPUT:
	// pumpkin
	// orange
	// strawberry
	// carrot
}
