package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg" // 通过 jpeg 包中的 init 函数注册解码器
	"image/png"
	"os"

	"github.com/nfnt/resize"
	drawx "golang.org/x/image/draw"
)

/*
开运 golang 图像库：
https://www.topgoer.com/%E5%BC%80%E6%BA%90/%E5%9B%BE%E7%89%87.html
[图像缩放算法小结](https://blog.csdn.net/allen_sdz/article/details/89166363)
*/

/*
源照片缩小为头像大小分割成 1 或者更 8*8 像素的格子，每个格子计算一个特征值（三通道）；
像素照片 1 张，4 张，9 张，16张，25 张（照片组）组合成照片组，计算特征值；采用图像压缩算法，计算一组图片压缩成1个像素（9 个像素点）点时的特征值。
在像素照片组中寻找和 源照片的格子 特征最接近的照片，填入对应格子中，组成最终的照片。

相似度：三通道的平方和？
按三通道，每个通道建立一个索引，每个通道超过 100 平方值就超过 10000 了，没必要参与计算了，用于减少计算量

一张图片压缩成 9、16、25个像素点？然后按像素点与原图进行比对？
用 Bilinear算法？

压缩成不同大小，然后对比多大时相似度比较好？做一个取舍

先用压缩到一个像素点做备选，在从备选中的图片放到后选择最相似的图片做填充

特征值计算参考图片压缩、插值算法，避免放大缩小时，失真太大
*/

func main() {
	input, _ := os.Open("pic/周鸿祎.jpeg")
	defer input.Close()
	img, str, err := image.Decode(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(img.Bounds(), ":", str)

	input, _ = os.Open("pic/雷军.jpeg")
	defer input.Close()
	img2, str, err := image.Decode(input)
	if err != nil {
		panic(err)
	}
	_ = img2
	// img = Draw([]image.Image{img, img2})

	img, err = ResizeImage(img, 512, 512)
	if err != nil {
		panic(err)
	}
	// img = Scale(img)
	saveImage(img, "test.jpeg")

	subImg, err := Sub(img, image.Rect(256, 256, 256+128, 256+128))
	if err != nil {
		panic(err)
	}
	saveImage(subImg, "aplit1.jpeg")
	subImgs, err := Split(img, 20, 20)
	if err != nil {
		panic(err)
	}
	for i, img := range subImgs {
		saveImage(img, fmt.Sprintf("aplit-%d.jpeg", i))
	}
}

/*
图片按比例缩小
有了上述的背景知识后再来实现图片按特定比例缩放就容易多了。下面的代码实现是把任意一张图片按照 16:9 的比例缩小，并且对比例失调的部分用白边补齐。
*/
// ResizeImage 图片按 16:9 缩小，白边补齐
func ResizeImage1(img image.Image) (newImg1 image.Image, err error) {
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
	m := resize.Resize(uint(newW), uint(newH), img, resize.Bicubic)

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
	// outBuffer := new(bytes.Buffer)
	// if err = jpeg.Encode(outBuffer, newImg, nil); err != nil {
	// 	return
	// }
	// out = outBuffer
	return newImg, nil
}

func NewBack(size image.Rectangle) *image.RGBA {
	// 生成背景图
	newImg := image.NewRGBA(size)

	// 背景图变成白色
	c := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for x := 0; x < size.Max.X; x++ {
		for y := 0; y < size.Max.Y; y++ {
			newImg.Set(x, y, c)
		}
	}
	return newImg
}
func ResizeImage(img image.Image, newW, newH int) (newImg1 image.Image, err error) {
	// 原图的尺寸
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	dW := float64(newW) / float64(width)
	dH := float64(newH) / float64(height)
	// 如果原图尺寸超过 比例 则缩小；不超过则保持原尺寸只进行白边补齐
	if dW > dH {
		width = int(float64(width) * dH)
		height = newH
	} else {
		height = int(float64(height) * dW)
		width = newW
	}

	// 图片缩小
	m := resize.Resize(uint(width), uint(height), img, resize.Bicubic)
	// return m, nil

	// 生成背景图
	newBack := NewBack(image.Rect(0, 0, newW, newH))

	// 在背景图上绘制缩小后的图片，实现补齐白边
	if dW < dH {
		draw.Draw(newBack,
			image.Rectangle{
				Min: image.Point{Y: (newH - m.Bounds().Dy()) / 2},
				Max: image.Point{X: newW, Y: (newH + m.Bounds().Dy()) / 2},
			},
			m, m.Bounds().Min, draw.Src)
	} else {
		draw.Draw(newBack,
			image.Rectangle{
				Min: image.Point{X: (newW - m.Bounds().Dx()) / 2},
				Max: image.Point{X: (newW + m.Bounds().Dx()) / 2, Y: newH},
			}, m, m.Bounds().Min, draw.Src)
	}

	// 编码生成图片
	// outBuffer := new(bytes.Buffer)
	// if err = jpeg.Encode(outBuffer, newImg, nil); err != nil {
	// 	return
	// }
	// out = outBuffer
	return newBack, nil
}

func Scale(src image.Image) (dstImg image.Image) {
	dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2)) // 缩放后的目标图片
	drawx.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), drawx.Over, nil)     // 使用 NearestNeighbor 算法进行伸缩

	return dst
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
func Draw(images []image.Image) (dstImg image.Image) {
	// jpeg 解码器返回的 image 对象（姑且称为对象）是只读的并不能在上面自由绘制，我们需要创建一个画布：
	width := 1080
	height := 1920
	dst := image.NewRGBA(image.Rect(0, 0, width, height))                                                  // 创建一块画布
	draw.Draw(dst, image.Rect(0, height/4, width/2, 3*height/4), images[0], image.Pt(0, 0), draw.Over)     // 绘制第一幅图
	draw.Draw(dst, image.Rect(width/2, height/4, width, 3*height/4), images[1], image.Pt(0, 0), draw.Over) // 绘制第二幅图

	return dst
}

func Sub(src image.Image, rect image.Rectangle) (subImg image.Image, err error) {
	if rgbImg, ok := src.(*image.YCbCr); ok {
		subImg = rgbImg.SubImage(rect).(*image.YCbCr) //图片裁剪x0 y0 x1 y1
	} else if rgbImg, ok := src.(*image.RGBA); ok {
		subImg = rgbImg.SubImage(rect).(*image.RGBA) //图片裁剪x0 y0 x1 y1
	} else if rgbImg, ok := src.(*image.NRGBA); ok {
		subImg = rgbImg.SubImage(rect).(*image.NRGBA) //图片裁剪x0 y0 x1 y1
	} else {
		return src, fmt.Errorf("图片解码失败, type:%+T", src)
	}
	return subImg, nil
}

func Split(src image.Image, nX, nY int) (dsts []image.Image, err error) {
	// 生成背景图
	w := src.Bounds().Dx() / nX
	h := src.Bounds().Dy() / nY
	for x := 0; x < nX; x++ {
		for y := 0; y < nY; y++ {
			xw, yh := x*w, y*h
			subImg, err := Sub(src, image.Rect(xw, yh, xw+w, yh+h))
			if err != nil {
				return dsts, err
			}
			dsts = append(dsts, subImg)
		}
	}
	return
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

func savePng(img image.Image, filename string) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, img, nil)
	if err != nil {
		return err
	}
	err = b.Flush()
	if err != nil {
		return err
	}
	return nil
}
