package user

import (
	//"github.com/nfnt/resize"
	"os"
	//"fmt"
	"image/jpeg"
	"image/png"
	"image"
	"errors"
	"github.com/disintegration/imaging"
)

func MiniMize(i_type string, ratio float64,  path string)(error){
	file, err := os.Open("./public/files/"+path)
	if err != nil {
		file.Close()
		return err
	}
	var image image.Image
	image = nil
	if i_type == "png"{
		image, err = png.Decode(file)
		if err != nil {
			file.Close()
			return err
		}
	}
	if i_type == "jpeg" || i_type == "jpg"{
		image, err = jpeg.Decode(file)
		if err != nil {
			file.Close()
			return err
		}
	}
	if image == nil{
		file.Close()
		return  errors.New("Failed type")
	}
	file.Close()
	g := image.Bounds()
	height := g.Dy()
	width := g.Dx()
	//fmt.Println("Width = ", width)
	//fmt.Println("Height = ", height)
	//fmt.Println("Ratio size = ", ratio_size)
	if width > 250 || height>180{
		width:= int(180* ratio)
		//newImage := resize.Resize(uint(width), 180, image, resize.Lanczos3)
		dstImage128 := imaging.Resize(image, width, 180, imaging.Lanczos)
		out, err := os.Create("./public/files/min/"+path)
		if err != nil {
			return err
		}
		defer out.Close()
		jpeg.Encode(out, dstImage128,nil)
	}
	return nil
}