package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg" // 通过 jpeg 包中的 init 函数注册解码器
	"image/png"
	_ "image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/image/draw"
)

/*
开运 golang 图像库：
https://www.topgoer.com/%E5%BC%80%E6%BA%90/%E5%9B%BE%E7%89%87.html
*/

func main() {
	input, _ := os.Open("pic/雷军.jpeg")
	defer input.Close()
	img, str, err := image.Decode(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(img.Bounds(), ":", str)
}
func Scale(src image.Image) (dst1 image.Image) {
	dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2)) // 缩放后的目标图片
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)       // 使用 NearestNeighbor 算法进行伸缩

	return dst
	// 	x/image 包中有四种缩放算法:
	// NearestNeighbor
	// ApproxBiLinear
	// BiLinear
	// CatmullRom
	// 也可以使用 github.com/nfnt/resize 包：
	// 第一个参数为缩放后的宽度
	// 第二个参数为缩放后的高度
	// 第三个参数为待缩放图片
	//第四个参数为使用哪种插值算法
	// resize.Resize(targetWidth, targetHeight, img, resize.NearestNeighbor)

	m := image.NewRGBA(image.Rect(0, 0, 640, 480))
	blue := color.RGBA{0, 0, 255, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

	return
}

/*
图片按比例缩小
有了上述的背景知识后再来实现图片按特定比例缩放就容易多了。下面的代码实现是把任意一张图片按照 16:9 的比例缩小，并且对比例失调的部分用白边补齐。
*/
// ResizeImage 图片按 16:9 缩小，白边补齐
func ResizeImage(file io.Reader) (out io.Reader, err error) {
	// file 是原图
	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	// 检查图片类型
	fileType := http.DetectContentType(buffer)
	var img image.Image
	if fileType == "image/png" {
		// decode jpeg into image.Image
		img, err = png.Decode(bytes.NewReader(buffer))
	} else if fileType == "image/jpeg" {
		img, err = jpeg.Decode(bytes.NewReader(buffer))
	} else {
		err = errors.New("invalid file type")
		return
	}
	if err != nil {
		return
	}

	// 原图的尺寸
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	var (
		required                 = 16 / 9
		ration                   = width / height
		newW, newH, backW, backH int
	)

	// 如果原图尺寸超过 1280 * 720 则缩小；不超过则保持原尺寸只进行白边补齐
	if ration > required {
		newH = 0
		if width > 1280 {
			newW = 1280
		} else {
			newW = width
		}
		backW = newW
		backH = backW * 9 / 16
	} else {
		newW = 0
		if newH > 720 {
			newH = 720
		} else {
			newH = height
		}
		backH = newH
		backW = backH * 16 / 9
	}

	// 图片缩小
	m := resize.Resize(uint(newW), uint(newH), img, resize.NearestNeighbor)

	// 生成背景图
	newImg := image.NewRGBA(image.Rect(0, 0, backW, backH))

	// 背景图变成白色
	c := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for x := 0; x < backW; x++ {
		for y := 0; y < backH; y++ {
			newImg.Set(x, y, c)
		}
	}

	// 在背景图上绘制缩小后的图片，实现补齐白边
	if ration > required {
		draw.Draw(newImg, image.Rectangle{Min: image.Point{Y: (backH - m.Bounds().Dy()) / 2}, Max: image.Point{X: backW, Y: (backH + m.Bounds().Dy()) / 2}}, m, m.Bounds().Min, draw.Src)
	} else {
		draw.Draw(newImg, image.Rectangle{Min: image.Point{X: (backW - m.Bounds().Dx()) / 2}, Max: image.Point{X: (backW + m.Bounds().Dx()) / 2, Y: backH}}, m, m.Bounds().Min, draw.Src)
	}

	// 编码生成图片
	outBuffer := new(bytes.Buffer)
	if err = jpeg.Encode(outBuffer, newImg, nil); err != nil {
		return
	}
	out = outBuffer
	return
}

/*
func Draw(dst Image, r image.Rectangle, src image.Image, sp image.Point, op Op) {
	DrawMask(dst, r, src, sp, nil, image.Point{}, op)
}
Draw 函数的参数如下：
	dst: 绘图的画布，只能是 *image.RGBA 或 *image.Paletted 类型
	r: dst 上绘图范围
	src: 要在画布上绘制的图像
	sp: src 的起始点，实际绘制的图像是img.SubImage(image.Rectangle{Min: sp, Max: src.Bounds().Max})。注意，src 为 SubImage 时，sp 应为 src.Bounds().Min
	op: porter-duff 混合方式, 具体介绍可以看下面的微软的参考资料。image/draw 库提供了 draw.Src 和 draw.Over 两种混合方式
		draw.Src: 将 src 覆盖在 dst 上
		draw.Over: src 在上，dst 在下按照 alpha 值进行混合。在图片完全不透明时，draw.Over 与 draw.Src 没有区别
Draw 不支持设置背景图片或者背景色，其实只要在画布最下层绘制一张和画布一样大的图片或纯色图片即可。
*/
func Draw(src image.Image) (dst image.Image) {
	m := image.NewRGBA(image.Rect(0, 0, 640, 480))
	blue := color.RGBA{0, 0, 255, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

	// jpeg 解码器返回的 image 对象（姑且称为对象）是只读的并不能在上面自由绘制，我们需要创建一个画布：
	width := 1080
	height := 1920
	dst = image.NewRGBA(image.Rect(0, 0, width, height))                                                   // 创建一块画布
	draw.Draw(dst, image.Rect(0, height/4, width/2, 3*height/4), images[0], image.Pt(0, 0), draw.Over)     // 绘制第一幅图
	draw.Draw(dst, image.Rect(width/2, height/4, width, 3*height/4), images[1], image.Pt(0, 0), draw.Over) // 绘制第二幅图

	return
}
func SubImage(src image.Image) (dst image.Image) {
	rgbImg := src.(*image.YCbCr)
	subImg := rgbImg.SubImage(image.Rect(0, 0, 200, 200)).(*image.YCbCr) //图片裁剪x0 y0 x1 y1

	return subImg
}

/*
func DrawMask(dst Image, r image.Rectangle, src image.Image, sp image.Point, mask image.Image, mp image.Point, op Op)
遮罩
DrawMask 函数可以在 src 上面一个遮罩，可以实现圆形图片、圆角等效果。
圆形图片
首先定义一个中心圆形不透明、边缘部分透明的 circle 类型，实现 image.Image 接口：
*/
// 圆形遮罩
type circle struct {
	p image.Point // 圆心位置
	r int         // 半径
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{A: 255} // 半径以内的图案设成完全不透明
	}
	return color.Alpha{}
}

func DrawMask(src image.Image) (dst image.Image) {
	// 使用 DrawMask 方法将其绘制出来：
	c := circle{p: image.Point{X: avatarRad, Y: avatarRad}, r: avatarRad}
	circleAvatar := image.NewRGBA(image.Rect(0, 0, avatarRad*2, avatarRad*2))                               // 准备画布
	draw.DrawMask(circleAvatar, circleAvatar.Bounds(), avatar, image.Point{}, &c, image.Point{}, draw.Over) // 使用 Over 模式进行混合
	return
}

/*
圆角
圆角的实现原理和圆形一样，改一下 At 的函数公式即可：
*/
type radius struct {
	p image.Point // 矩形右下角位置
	r int
}

func (c *radius) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *radius) Bounds() image.Rectangle {
	return image.Rect(0, 0, c.p.X, c.p.Y)
}

// 对每个像素点进行色值设置，分别处理矩形的四个角，在四个角的内切圆的外侧，色值设置为全透明，其他区域不透明
func (c *radius) At(x, y int) color.Color {
	var xx, yy, rr float64
	var inArea bool
	// left up
	if x <= c.r && y <= c.r {
		xx, yy, rr = float64(c.r-x)+0.5, float64(y-c.r)+0.5, float64(c.r)
		inArea = true
	}
	// right up
	if x >= (c.p.X-c.r) && y <= c.r {
		xx, yy, rr = float64(x-(c.p.X-c.r))+0.5, float64(y-c.r)+0.5, float64(c.r)
		inArea = true
	}
	// left bottom
	if x <= c.r && y >= (c.p.Y-c.r) {
		xx, yy, rr = float64(c.r-x)+0.5, float64(y-(c.p.Y-c.r))+0.5, float64(c.r)
		inArea = true
	}
	// right bottom
	if x >= (c.p.X-c.r) && y >= (c.p.Y-c.r) {
		xx, yy, rr = float64(x-(c.p.X-c.r))+0.5, float64(y-(c.p.Y-c.r))+0.5, float64(c.r)
		inArea = true
	}

	if inArea && xx*xx+yy*yy >= rr*rr {
		return color.Alpha{}
	}
	return color.Alpha{A: 255}
}

/*
添加文字
github.com/golang/freetype 库可以用来在图片上绘制文字:
*/

func freetype() {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	ttfBytes, err := ioutil.ReadFile(fontSource) // 读取 ttf 文件
	if err != nil {
		return err
	}

	font, err := freetype.ParseFont(ttfBytes)
	if err != nil {
		return err
	}

	fc := freetype.NewContext()
	fc.SetDPI(72) // 每英寸的分辨率
	fc.SetFont(font)
	fc.SetFontSize(size)
	fc.SetClip(img.Bounds())
	fc.SetDst(img)
	fc.SetSrc(image.Black) // 设置绘制操作的源图像，通常使用纯色图片 image.Uniform

	_, err = fc.DrawString("hello world", freetype.Pt(0, 0))
	if err != nil {
		return err
	}
}

func saveImage(img image.Image, filename string) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = jpeg.Encode(b, img, nil)
	if err != nil {
		return err
	}
	err = b.Flush()
	if err != nil {
		return err
	}
	return nil
}
