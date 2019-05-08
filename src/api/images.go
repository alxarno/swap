package api

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/swap-messenger/swap/settings"
)

func compressionImage(iType string, ratio float64, path string) error {
	iType = strings.ToLower(iType)
	file, err := os.Open(settings.ServiceSettings.Backend.FilesPath + path)
	defer file.Close()
	if err != nil {
		log.Println(err, 0)
		return err
	}
	var nowImage image.Image
	nowImage = nil
	if iType == "png" {
		nowImage, err = png.Decode(file)
		if err != nil {
			log.Println(err, 1)
			return err
		}
	}
	if iType == "jpeg" || iType == "jpg" {
		nowImage, err = jpeg.Decode(file)
		if err != nil {
			return errors.New("JPG/JPEG encode error:" + err.Error())
		}
	}
	if nowImage == nil {
		return errors.New("Failed type error: " + iType)
	}
	file.Close()
	g := nowImage.Bounds()
	height := g.Dy()
	width := g.Dx()
	out, err := os.Create(settings.ServiceSettings.Backend.FilesPath + "min//" + path)
	if err != nil {
		log.Println(err, 3)
		return errors.New("Failed creating file:" + err.Error())
	}
	defer out.Close()
	if width > 500 || height > 360 {
		height = 360
		width = int(float64(height) * ratio)
		dstImage128 := imaging.Resize(nowImage, width, height, imaging.Lanczos)
		dstImage := dstImage128

		if iType == "png" {
			png.Encode(out, dstImage)
		}
		if iType == "jpeg" || iType == "jpg" {
			Options := jpeg.Options{Quality: 70}
			jpeg.Encode(out, dstImage, &Options)
		}

	} else {
		if iType == "png" {
			png.Encode(out, nowImage)
		}
		if iType == "jpeg" || iType == "jpg" {
			Options := jpeg.Options{Quality: 70}
			jpeg.Encode(out, nowImage, &Options)
		}
	}
	return nil
}
