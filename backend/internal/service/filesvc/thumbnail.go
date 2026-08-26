package filesvc

import (
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	thumbFolderName  = "thumbnails"
	jpegThumbQuality = 82
	thumbFileMode    = 0o640
	thumbDirMode     = 0o750
	maxImageWidth    = 8000
	maxImageHeight   = 8000
)

var (
	// MaxThumbWidth and MaxThumbHeight cap the thumbnail size; images smaller
	// than these dimensions are never upscaled.
	MaxThumbWidth  = 300
	MaxThumbHeight = 300
)

// ThumbnailPath returns where the pre-generated thumbnail of the given file id
// lives on disk. It is deterministically derived from the storage root, so it
// needs no database column.
func (s *Service) ThumbnailPath(id string) string {
	return filepath.Join(s.storagePath, thumbFolderName, id)
}

// generateThumbnail decodes the image stored at sourcePath, writes a resized
// copy to destinationPath (created if its directory is missing) and keeps the
// source untouched. The output format matches the input format: JPEG stays
// JPEG (quality 82), PNG stays PNG and GIF becomes a static GIF of its first
// frame. Corrupted or unsupported images return an error without leaving a
// partial file behind.
func generateThumbnail(sourcePath, destinationPath string) error {
	if err := os.MkdirAll(filepath.Dir(destinationPath), thumbDirMode); err != nil {
		return err
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	contentType, err := detectFileType(source)
	if err != nil {
		return err
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err := validateImageDimensions(source); err != nil {
		return err
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return err
	}

	decoded, err := decodeImage(contentType, source)
	if err != nil {
		return err
	}

	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, thumbFileMode)
	if err != nil {
		return err
	}
	if err := encodeImage(destination, contentType, fitThumbnail(decoded)); err != nil {
		destination.Close()
		_ = os.Remove(destinationPath)
		return err
	}
	if err := destination.Close(); err != nil {
		_ = os.Remove(destinationPath)
		return err
	}
	return nil
}

func detectFileType(source *os.File) (string, error) {
	buffer := make([]byte, 512)
	count, err := source.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	contentType := http.DetectContentType(buffer[:count])
	if !allowedImageType(contentType) {
		return "", ErrInvalidImage
	}
	return contentType, nil
}

func validateImageDimensions(source *os.File) error {
	config, _, err := image.DecodeConfig(source)
	if err != nil {
		return ErrInvalidImage
	}
	if config.Width > maxImageWidth || config.Height > maxImageHeight {
		return ErrInvalidImage
	}
	return nil
}

func decodeImage(contentType string, source *os.File) (image.Image, error) {
	switch contentType {
	case "image/jpeg":
		return jpeg.Decode(source)
	case "image/png":
		return png.Decode(source)
	case "image/gif":
		return gif.Decode(source)
	default:
		return nil, ErrInvalidImage
	}
}

func encodeImage(destination *os.File, contentType string, img image.Image) error {
	switch contentType {
	case "image/jpeg":
		return jpeg.Encode(destination, img, &jpeg.Options{Quality: jpegThumbQuality})
	case "image/png":
		return png.Encode(destination, img)
	case "image/gif":
		return gif.Encode(destination, img, nil)
	default:
		return ErrInvalidImage
	}
}

// fitThumbnail scales the image down so it fits inside MaxThumbWidth x
// MaxThumbHeight while preserving the aspect ratio. Images already within the
// limits are returned unchanged (never upscaled).
func fitThumbnail(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= MaxThumbWidth && height <= MaxThumbHeight {
		return img
	}
	scaleX := float64(MaxThumbWidth) / float64(width)
	scaleY := float64(MaxThumbHeight) / float64(height)
	targetWidth := max(int(float64(width)*min(scaleX, scaleY)), 1)
	targetHeight := max(int(float64(height)*min(scaleX, scaleY)), 1)
	return scaleImage(img, targetWidth, targetHeight)
}

// scaleImage downscales using an area-average (box filter): every destination
// pixel is the average of all source pixels that map onto it. Channels are
// alpha-premultiplied, which keeps transparent edges clean.
func scaleImage(source image.Image, width, height int) image.Image {
	bounds := source.Bounds()
	target := image.NewRGBA(image.Rect(0, 0, width, height))
	scaleX := float64(bounds.Dx()) / float64(width)
	scaleY := float64(bounds.Dy()) / float64(height)

	for y := range height {
		startY := float64(y) * scaleY
		endY := float64(y+1) * scaleY
		for x := range width {
			startX := float64(x) * scaleX
			endX := float64(x+1) * scaleX

			var sumR, sumG, sumB, sumA, samples float64
			for py := int(startY); float64(py) < endY && py < bounds.Dy(); py++ {
				for px := int(startX); float64(px) < endX && px < bounds.Dx(); px++ {
					r, g, b, a := source.At(bounds.Min.X+px, bounds.Min.Y+py).RGBA()
					sumR += float64(r)
					sumG += float64(g)
					sumB += float64(b)
					sumA += float64(a)
					samples++
				}
			}
			target.SetRGBA64(x, y, color.RGBA64{
				R: uint16(sumR / samples),
				G: uint16(sumG / samples),
				B: uint16(sumB / samples),
				A: uint16(sumA / samples),
			})
		}
	}
	return target
}
