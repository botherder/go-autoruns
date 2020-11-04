package main

import (
	"fmt"
	autoruns "github.com/botherder/go-autoruns"
)

func main() {
	autorunList := autoruns.Autoruns()

	for _, autorun := range autorunList {
		fmt.Println(autorun.Type)
		fmt.Println(autorun.Location)
		fmt.Println(autorun.ImagePath)
		fmt.Println(autorun.ImageName)
		fmt.Println(autorun.Arguments)
		fmt.Println(autorun.Entry)
		fmt.Println(autorun.LaunchString)
		fmt.Println("")
	}
}
