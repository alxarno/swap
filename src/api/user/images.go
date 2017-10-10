package user

import (
	//"github.com/nfnt/resize"
	"os"
	//"fmt"
	"image/jpeg"
	"image/png"
	"image"
	"errors"
	//"github.com/disintegration/imaging"
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
	out, err := os.Create("./public/files/min/"+path)
	if err != nil {
		return err
	}
	defer out.Close()
	if width > 500 || height>360{
		height =360
		width =int(float64(height)*ratio)
		//newImage := resize.Resize(uint(width), 180, image, resize.Lanczos3)
		//imaging.
		dstImage128 := imaging.Resize(image, width, height, imaging.Lanczos)
		//dstImage128:= imaging.CropAnchor(image, width, height, imaging.Center)
		//dstImage := imaging.Blur(dstImage128, 1.3)
		dstImage:= dstImage128

		//Options := struct {Quality int}{70}
		if i_type == "png"{
			//Options:=png.Options{70}
			png.Encode(out, dstImage)
		}
		if i_type == "jpeg" || i_type == "jpg"{
			Options:=jpeg.Options{70}
			jpeg.Encode(out, dstImage,&Options)
		}

	}else{
		if i_type == "png"{
			//Options:=png.Options{70}
			png.Encode(out, image)
		}
		if i_type == "jpeg" || i_type == "jpg"{
			Options:=jpeg.Options{70}
			jpeg.Encode(out, image, &Options)
		}
	}
	return nil
}