package main

import (
	"fmt"
	"github.com/reaburoa/utils/picture"
)

func main() {
	imgFile := "./DSC01225.JPG"
	er := picture.NewPicture(imgFile).Crop(500, 0, 800, 1500, "DSC01225_subimg.jpg")
	if er != nil {
		fmt.Println("Crop File Failed", er.Error())
	}
}
