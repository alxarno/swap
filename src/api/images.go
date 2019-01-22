package api

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

func compressionImage(iType string, ratio float64,  path string)(error){
	file, err := os.Open("./public/files/"+path)
	defer file.Close();if err != nil {
		return err
	}
	var nowImage image.Image
	nowImage = nil
	if iType == "png"{
		nowImage, err = png.Decode(file);if err != nil {
			return err
		}
	}
	if iType == "jpeg" || iType == "jpg"{
		nowImage, err = jpeg.Decode(file);if err != nil {
			return err
		}
	}
	if nowImage == nil{
		return  errors.New("failed type")
	}
	file.Close()
	g := nowImage.Bounds()
	height := g.Dy()
	width := g.Dx()
	//fmt.Println("Width = ", width)
	//fmt.Println("Height = ", height)
	//fmt.Println("Ratio size = ", ratio_size)
	out, err := os.Create("./public/files/min/"+path);if err != nil {
		return err
	}
	defer out.Close()
	if width > 500 || height>360{
		height =360
		width =int(float64(height)*ratio)
		//newImage := resize.Resize(uint(width), 180, nowImage, resize.Lanczos3)
		//imaging.
		dstImage128 := imaging.Resize(nowImage, width, height, imaging.Lanczos)
		//dstImage128:= imaging.CropAnchor(nowImage, width, height, imaging.Center)
		//dstImage := imaging.Blur(dstImage128, 1.3)
		dstImage:= dstImage128

		//Options := struct {Quality int}{70}
		if iType == "png"{
			//Options:=png.Options{70}
			png.Encode(out, dstImage)
		}
		if iType == "jpeg" || iType == "jpg"{
			Options:=jpeg.Options{Quality:70}
			jpeg.Encode(out, dstImage,&Options)
		}

	}else{
		if iType == "png"{
			//Options:=png.Options{70}
			png.Encode(out, nowImage)
		}
		if iType == "jpeg" || iType == "jpg"{
			Options:=jpeg.Options{Quality:70}
			jpeg.Encode(out, nowImage, &Options)
		}
	}
	return nil
}