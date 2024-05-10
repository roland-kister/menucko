package imageocr

import (
	"bytes"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
	"github.com/otiai10/gosseract/v2"
)

type ImageOcr interface {
	ParseJpegText(imgBytes []byte) (string, error)
}

type DevImageOcr struct {
	ImgText string
}

func (imageOcr DevImageOcr) ParseJpegText(imgBytes []byte) (string, error) {
	return imageOcr.ImgText, nil
}

type ProdImageOcr struct{}

func (ProdImageOcr) ParseJpegText(imgBytes []byte) (string, error) {
	img, err := jpeg.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return "", err
	}

	img = imaging.Resize(img, img.Bounds().Max.X*4, img.Bounds().Max.Y*4, imaging.Lanczos)
	img = imaging.AdjustContrast(img, 10)
	img = imaging.Sharpen(img, 4)

	var imgBuf bytes.Buffer
	if err = png.Encode(&imgBuf, img); err != nil {
		return "", err
	}

	client := gosseract.NewClient()
	defer client.Close()

	if err = client.SetLanguage("slk"); err != nil {
		return "", err
	}

	if err = client.SetPageSegMode(gosseract.PSM_SINGLE_BLOCK); err != nil {
		return "", err
	}

	if err = client.SetImageFromBytes(imgBuf.Bytes()); err != nil {
		return "", err
	}

	return client.Text()
}
