package picture

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/golang/freetype"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FileType string

type Picture struct {
	// 图片原始路径
	imgFile string
	// 图片名，不包含扩展名称
	imgName string
	// 图片类型
	imgFileType FileType
	// 图片质量，0-100间，数字越大质量越高
	quality int
	// 图片路径
	imgFilePath string
}

const (
	FileTypeJPG FileType = "jpg"
	FileTypePNG FileType = "png"
	FileTypeGIF FileType = "gif"
	FileTypeTIF FileType = "tif"
	FileTypeBMP FileType = "bmp"
)

func NewPicture(img string) *Picture {
	fInfo, err := os.Stat(img)
	if err != nil {
		panic("Check File Error:" + err.Error())
	}
	if fInfo.IsDir() {
		panic(img + " Is Not A Image")
	}
	picIns := &Picture{
		imgFile: img,
	}
	filePath, _ := picIns.getFilePath()
	picIns.imgFilePath = filePath
	fSrc, err := ioutil.ReadFile(picIns.imgFile)
	if err != nil {
		panic("Check File Error:" + err.Error())
	}
	name, imgType := picIns.getFileInfo(fSrc)
	if imgType == "" || (imgType != FileTypeJPG && imgType != FileTypePNG && imgType != FileTypeGIF) {
		panic("Not Support File Type")
	}
	picIns.imgFileType = imgType
	picIns.imgName = name

	return picIns
}

var (
	fileByteMap = map[string]FileType{
		"ffd8ffe000104a464946": FileTypeJPG, // JPEG (jpg)
		"89504e470d0a1a0a0000": FileTypePNG, // PNG (png)
		"47494638396126026f01": FileTypeGIF, // GIF (gif)
		"49492a00227105008037": FileTypeTIF, // TIFF (tif)
		"424d228c010000000000": FileTypeBMP, // 16色位图(bmp)
		"424d8240090000000000": FileTypeBMP, // 24位位图(bmp)
		"424d8e1b030000000000": FileTypeBMP,
	}

	fileExtMap = map[string]FileType{
		"jpg":  FileTypeJPG, // JPEG (jpg)
		"jpeg": FileTypePNG, // PNG (png)
		"gif":  FileTypeGIF, // GIF (gif)
		"tif":  FileTypeTIF, // TIFF (tif)
		"bmp":  FileTypeBMP, // 16色位图(bmp)
	}
)

func (p *Picture) bytesToHexString(src []byte) string {
	if src == nil || len(src) <= 0 {
		return ""
	}
	res := bytes.Buffer{}
	tmp := make([]byte, 0)
	for i := 0; i < len(src); i++ {
		v := src[i] & 0xFF
		hv := hex.EncodeToString(append(tmp, v))
		if len(hv) < 2 {
			res.WriteString(strconv.FormatInt(int64(0), 10))
		}
		res.WriteString(hv)
	}
	return res.String()
}

func (p *Picture) getFileInfo(fileSrc []byte) (string, FileType) {
	name, ext := p.getFileName()
	fileCode := p.bytesToHexString(fileSrc)
	for key, value := range fileByteMap {
		if strings.HasPrefix(fileCode, strings.ToLower(key)) || strings.HasPrefix(key, strings.ToLower(fileCode)) {
			return name, value
		}
	}
	return name, ext
}

func (p *Picture) getFileName() (string, FileType) {
	namesArr := strings.Split(strings.ToLower(filepath.Base(p.imgFile)), ".")
	filename := strings.Join(namesArr[:len(namesArr)-1], ".")
	ext, ok := fileExtMap[namesArr[len(namesArr)-1]]
	if ok {
		return filename, ext
	}
	return filename, FileType(namesArr[len(namesArr)-1])
}

func (p *Picture) getFilePath() (string, string) {
	return filepath.Split(p.imgFile)
}

func (p *Picture) SetQuality(quality int) {
	p.quality = quality
}

func (p *Picture) openImage() (*os.File, error) {
	fp, err := os.Open(p.imgFile)
	if err != nil {
		return nil, err
	}

	return fp, nil
}

func (p *Picture) decodeFile() (image.Image, error) {
	fp, err := p.openImage()
	if err != nil {
		return nil, err
	}
	var (
		dcode image.Image
		er    error
	)
	switch p.imgFileType {
	case FileTypeJPG:
		dcode, er = jpeg.Decode(fp)
	case FileTypePNG:
		dcode, er = png.Decode(fp)
	case FileTypeGIF:
		dcode, er = gif.Decode(fp)
	}
	if er != nil {
		return nil, er
	}
	return dcode, nil
}

func (p *Picture) Crop(width, height, endWidth, endHeight int, subImgPath string) error {
	imgPal, err := p.decodeFile()
	if err != nil {
		return err
	}
	var subImg image.Image
	quality := 100
	switch p.imgFileType {
	case FileTypeJPG:
		subImg = imgPal.(*image.YCbCr).SubImage(image.Rect(width, height, endWidth, endHeight))
	case FileTypePNG:
		subImg = imgPal.(*image.Paletted).SubImage(image.Rect(width, height, endWidth, endHeight))
	}
	err = p.save(subImgPath, subImg, quality)
	if err != nil {
		return err
	}

	return nil
}

func (p *Picture) Compress(quality int, compressFile string) error {
	imgPal, err := p.decodeFile()
	if err != nil {
		return err
	}
	if p.imgFileType == FileTypePNG {
		p.imgFileType = FileTypeJPG
	}
	err = p.save(compressFile, imgPal, quality)
	return err
}

func (p *Picture) setFreeType(fontFile string, rgba *image.RGBA, dpi, fontSize int) *freetype.Context {
	freetypeCtx := freetype.NewContext()
	font, _ := ioutil.ReadFile(fontFile)
	tf, _ := freetype.ParseFont(font)
	freetypeCtx.SetFont(tf)
	freetypeCtx.SetDst(rgba)
	freetypeCtx.SetDPI(float64(dpi))
	freetypeCtx.SetFontSize(float64(fontSize))
	freetypeCtx.SetClip(rgba.Bounds())
	return freetypeCtx
}

func (p *Picture) Watermark(mark, fontFile string, dpi, fontSize, waterX, waterY int) error {
	imgPal, err := p.decodeFile()
	if err != nil {
		return err
	}

	rgba := image.NewRGBA(imgPal.Bounds())
	draw.Draw(rgba, rgba.Bounds(), imgPal, image.Point{}, draw.Src)

	freetypeCtx := p.setFreeType(fontFile, rgba, dpi, fontSize)

	r := uint8(211)
	g := uint8(211)
	b := uint8(211)
	freetypeCtx.SetSrc(image.NewUniform(color.RGBA{r, g, b, 255}))

	if waterX != 0 && waterY != 0 {
		pt := freetype.Pt(waterX, waterY)
		_, _ = freetypeCtx.DrawString(mark, pt)
	} else {
		x, y := 0, 0
		for y < imgPal.Bounds().Max.Y {
			xStep := 0
			for x < imgPal.Bounds().Max.X {
				pt := freetype.Pt(x, y)
				offset, _ := freetypeCtx.DrawString(mark, pt)
				if x == 0 {
					xStep = offset.X.Ceil()
				}
				x += xStep + 5
			}
			y += fontSize/5 + 15
			x = 0
		}
	}
	return p.save(p.imgFilePath+"/"+p.imgName+"_watermark."+string(p.imgFileType), rgba, 100)
}

func (p *Picture) GenWatermarkImg(origImg image.Image, mark, fontFile string, dpi, fontSize int) error {
	rgba := image.NewRGBA(origImg.Bounds())
	draw.Draw(rgba, rgba.Bounds(), origImg, image.Point{}, draw.Src)
	for x := 0; x < origImg.Bounds().Dx(); x++ {
		for y := 0; y < origImg.Bounds().Dy(); y++ {
			rgba.Set(x, y, image.NewUniform(color.RGBA{255, 255, 255, 255}))
		}
	}

	freetypeCtx := freetype.NewContext()
	font, _ := ioutil.ReadFile(fontFile)
	tf, _ := freetype.ParseFont(font)
	freetypeCtx.SetFont(tf)
	freetypeCtx.SetDst(rgba)
	freetypeCtx.SetDPI(float64(dpi))
	freetypeCtx.SetClip(rgba.Bounds())

	r := uint8(211)
	g := uint8(211)
	b := uint8(211)
	freetypeCtx.SetSrc(image.NewUniform(color.RGBA{r, g, b, 255}))
	freetypeCtx.SetFontSize(float64(fontSize))
	x, y := 0, 0
	for y < origImg.Bounds().Max.Y {
		xStep := 0
		for x < origImg.Bounds().Max.X {
			pt := freetype.Pt(x, y)
			offset, _ := freetypeCtx.DrawString(mark, pt)
			fmt.Println(offset.Y.Round())
			if x == 0 {
				xStep = offset.X.Ceil()
			}
			x += xStep + 5
		}
		y += fontSize/5 + 15
		x = 0
	}
	return p.save("watermark."+string(p.imgFileType), rgba, 100)
}

func (p *Picture) G(mark, fontFile string, dpi, fontSize int) error {
	imgPal, err := p.decodeFile()
	if err != nil {
		return err
	}

	er := p.GenWatermarkImg(imgPal, mark, fontFile, dpi, fontSize)
	if er != nil {
		return er
	}

	fp, err := os.Open("watermark.png")
	if err != nil {
		return err
	}
	dcode, er := png.Decode(fp)
	if er != nil {
		return er
	}

	rgba := image.NewRGBA(imgPal.Bounds())
	mask := image.NewUniform(color.Alpha{128})
	draw.DrawMask(rgba, dcode.Bounds(), dcode, image.Point{}, mask, image.Point{}, draw.Over)
	return p.save("ddd.png", rgba, 100)
}

func (p *Picture) save(imgPath string, pal image.Image, quality int) error {
	subFp, err := os.Create(imgPath)
	if err != nil {
		return nil
	}
	defer subFp.Close()
	var er error
	switch p.imgFileType {
	case FileTypeJPG:
		if quality <= 0 {
			quality = 10
		} else if quality >= 100 {
			quality = 100
		}
		er = jpeg.Encode(subFp, pal, &jpeg.Options{Quality: quality})
	case FileTypePNG:
		er = png.Encode(subFp, pal)
	case FileTypeGIF:
		er = gif.Encode(subFp, pal, &gif.Options{NumColors: 255})
	}
	if er != nil {
		return er
	}
	return nil
}
