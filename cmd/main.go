package main

import (
	"fmt"
	"github.com/botherder/go-autoruns"
)

func main() {
	autoruns := autoruns.Autoruns()

	for _, autorun := range autoruns {
		fmt.Println(autorun.Type)
		fmt.Println(autorun.Location)
		fmt.Println(autorun.ImagePath)
		fmt.Println(autorun.ImageName)
		fmt.Println(autorun.Arguments)
		fmt.Println("")
	}
}
