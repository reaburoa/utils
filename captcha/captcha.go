package captcha

import (
	"encoding/base64"
	"fmt"
	"github.com/golang/freetype"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Captcha struct {
	width, length int     // 图形验证码的宽度长度
	fontSize      float64 // 图形验证码字符大小
	font          string  // 图形验证码文字字体
	fontDpi       float64 // 图形验证码清晰度
	charTotal     int     // 图形验证码字符个数
	captchaMode   int     // 图形验证码字符模式
	rgba          *image.RGBA
	gc            *draw2dimg.GraphicContext
	freetypeCtx   *freetype.Context
}

const (
	numbers      = "0123456789"
	alphabet     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	mathOperator = "+-*"
)

const (
	CaptchaModeNumber = iota
	CaptchaModeString
	CaptchaModeMix
	CaptchaModeMathExpression
	CaptchaModeOther
)

type ImageType int

const (
	ImageTypeJpg ImageType = iota
	ImageTypePng
)

// 获取验证码实例
func NewCaptcha(width, length, charTotal, captchaMode int, fontSize float64, font string) *Captcha {
	captchaIns := &Captcha{
		width:       width,
		length:      length,
		fontSize:    fontSize,
		font:        font,
		charTotal:   charTotal,
		captchaMode: captchaMode,
	}
	captchaIns.rgba = captchaIns.initCanvas()
	captchaIns.gc = draw2dimg.NewGraphicContext(captchaIns.rgba)
	captchaIns.freetypeCtx = captchaIns.setFreeType()
	return captchaIns
}

func (c *Captcha) initCanvas() *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, c.length, c.width))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	for x := 0; x < c.length; x++ {
		for y := 0; y < c.width; y++ {
			rgba.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}

	return rgba
}

func (c *Captcha) randInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max)
}

func (c *Captcha) randNumbers() string {
	i := 0
	no := make([]string, 0, c.charTotal)
	max := len(numbers)
	for i < c.charTotal {
		n := c.randInt(0, max)
		no = append(no, string(numbers[n]))
		i++
	}

	return strings.Join(no, "")
}

func (c *Captcha) randChars() string {
	i := 0
	chars := make([]string, 0, c.charTotal)
	max := len(alphabet)
	for i < c.charTotal {
		n := c.randInt(0, max)
		chars = append(chars, string(alphabet[n]))
		i++
	}

	return strings.Join(chars, "")
}

func (c *Captcha) randMix() string {
	i := 0
	chars := make([]string, 0, c.charTotal)
	mixString := fmt.Sprintf("%s%s", alphabet, numbers)
	max := len(mixString)
	for i < c.charTotal {
		n := c.randInt(0, max)
		chars = append(chars, string(mixString[n]))
		i++
	}

	return strings.Join(chars, "")
}

func (c *Captcha) randMathExpression() (string, int) {
	n := c.randInt(0, len(mathOperator))
	leftNumber := c.randInt(0, 10)
	rightNumber := c.randInt(1, 10)

	var (
		res int
		op  string
	)
	switch mathOperator[n] {
	case '+':
		res = leftNumber + rightNumber
		op = "+"
	case '-':
		res = leftNumber - rightNumber
		op = "-"
	case '*':
		res = leftNumber * rightNumber
		op = "*"
	case '/':
		res = leftNumber / rightNumber
		op = "/"
	}

	return fmt.Sprintf("%d%s%d=?", leftNumber, op, rightNumber), res
}

func (c *Captcha) genCaptcha() (string, int) {
	var (
		res        int
		captchaStr string
	)
	switch c.captchaMode {
	case CaptchaModeNumber:
		captchaStr = c.randNumbers()
	case CaptchaModeString:
		captchaStr = c.randChars()
	case CaptchaModeMix:
		captchaStr = c.randMix()
	case CaptchaModeMathExpression:
		captchaStr, res = c.randMathExpression()
	case CaptchaModeOther:
	}

	return captchaStr, res
}

func (c *Captcha) mixLine() {
	gc := draw2dimg.NewGraphicContext(c.rgba)
	for i := 0; i < 10; i++ {
		gc.SetLineWidth(1)
		r := uint8(c.randInt(0, 255))
		g := uint8(c.randInt(0, 255))
		b := uint8(c.randInt(0, 255))
		gc.SetStrokeColor(color.RGBA{r, g, b, 255})

		startX := c.randInt(0, c.length+10)
		startY := c.randInt(0, c.width+10)

		lineX := c.randInt(0, c.length+10)
		lineY := c.randInt(0, c.width+10)
		gc.MoveTo(float64(startX), float64(startY))
		gc.LineTo(float64(lineX), float64(lineY))

		gc.Stroke()
	}
}

func (c *Captcha) mixSinLine() {
	gc := draw2dimg.NewGraphicContext(c.rgba)
	for i := 0; i < 1; i++ {
		gc.SetLineWidth(float64(c.randInt(2, 4)))
		r := uint8(c.randInt(0, 255))
		g := uint8(c.randInt(0, 255))
		b := uint8(c.randInt(0, 255))
		gc.SetStrokeColor(color.RGBA{r, g, b, 255})

		h1 := c.randInt(-12, 12)
		h2 := c.randInt(-1, 1)
		w2 := c.randInt(5, 20)
		h3 := c.randInt(5, 10)

		h := float64(c.width)
		w := float64(c.length)

		for j := -w / 2; j < w/2; j += 0.1 {
			y := h/float64(h3)*math.Sin(j/float64(w2)) + h/2 + float64(h1)
			gc.LineTo(j+w/2, y)
			if h2 == 0 {
				gc.LineTo(j+w/2, y+float64(h2))
			}
		}

		gc.Stroke()
	}
}

func (c *Captcha) mixPoint() {
	for i := 0; i < 150; i++ {
		c.gc.SetLineWidth(1)
		r := uint8(c.randInt(0, 255))
		g := uint8(c.randInt(0, 255))
		b := uint8(c.randInt(0, 255))
		c.gc.SetStrokeColor(color.RGBA{r, g, b, 255})

		startX := c.randInt(0, c.length+10)
		startY := c.randInt(0, c.width+5)
		c.gc.MoveTo(float64(startX), float64(startY))
		c.gc.LineTo(float64(startX+c.randInt(0, 2)), float64(startY+c.randInt(0, 1)))

		c.gc.Stroke()
	}
}

func (c *Captcha) SetFontDPI(dpi float64) {
	c.freetypeCtx.SetDPI(dpi)
}

func (c *Captcha) setFreeType() *freetype.Context {
	freetypeCtx := freetype.NewContext()
	font, _ := ioutil.ReadFile(c.font)
	tf, _ := freetype.ParseFont(font)
	freetypeCtx.SetFont(tf)
	freetypeCtx.SetDst(c.rgba)
	freetypeCtx.SetClip(c.rgba.Bounds())
	return freetypeCtx
}

func (c *Captcha) GenCode() (string, int, error) {
	code, res := c.genCaptcha()
	c.mixLine()
	c.mixSinLine()
	c.mixPoint()
	avg := (c.length - 20) / len(code)
	avgY := c.width/2 + int(c.fontSize)/2
	for i, v := range code {
		x := i*avg + avg/2
		err := c.writeChar(x, avgY, string(v))
		if err != nil {
			return "", 0, err
		}
	}
	defer c.gc.Close()
	defer c.gc.FillStroke()

	return code, res, nil
}

func (c *Captcha) writeChar(x, y int, str string) error {
	r := uint8(c.randInt(0, 250))
	g := uint8(c.randInt(0, 250))
	b := uint8(c.randInt(0, 250))
	c.freetypeCtx.SetSrc(image.NewUniform(color.RGBA{r, g, b, 255}))
	c.freetypeCtx.SetFontSize(float64(c.randInt(int(c.fontSize-2), int(c.fontSize+5))))
	pt := freetype.Pt(x, y)
	_, err := c.freetypeCtx.DrawString(str, pt)

	return err
}

func (c *Captcha) SaveJPG(filename string, quality int) error {
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	er := jpeg.Encode(fp, c.rgba, &jpeg.Options{Quality: quality})
	if er != nil {
		return err
	}

	return nil
}

func (c *Captcha) SavePNG(filename string) error {
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	er := png.Encode(fp, c.rgba)
	if er != nil {
		return err
	}

	return nil
}

func (c *Captcha) SaveBase64(imgType ImageType) (string, error) {
	var (
		fileName     = ""
		base64Header = ""
		er           error
	)
	switch imgType {
	case ImageTypeJpg:
		fileName = "tmp.jpg"
		base64Header = "data:image/jpg;base64,"
		er = c.SaveJPG(fileName, 60)
	case ImageTypePng:
		fileName = "tmp.png"
		base64Header = "data:image/png;base64,"
		er = c.SavePNG(fileName)
	}
	if er != nil {
		return "", er
	}
	all, _ := ioutil.ReadFile(fileName)

	str := base64.StdEncoding.EncodeToString(all)

	defer os.Remove(fileName)

	return fmt.Sprintf("%s%s", base64Header, str), nil
}
