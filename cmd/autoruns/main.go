package main

import (
	"fmt"
	"github.com/botherder/go-autoruns/v2"
)

func main() {
	allAutoruns := autoruns.GetAllAutoruns()

	for _, autorun := range allAutoruns {
		fmt.Println(autorun.Type)
		fmt.Println(autorun.Location)
		fmt.Println(autorun.ImagePath)
		fmt.Println(autorun.ImageName)
		fmt.Println(autorun.Arguments)
		fmt.Println("")
	}
}
