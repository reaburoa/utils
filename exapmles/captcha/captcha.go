package main

import (
	"fmt"
	"github.com/reaburoa/utils/captcha"
)

func main() {
	cc := captcha.NewCaptcha(60, 180, 4, captcha.CaptchaModeMix, 20, "./font/RitaSmith.ttf")
	cc.SetFontDPI(90)
	code, res, err := cc.GenCode()
	fmt.Println("genCode", code, res, err)
	er := cc.SaveJPG("captcha.jpg", 80)
	fmt.Println("SaveImage", er)
}
