package screenshots

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"text/template"

	"github.com/jtolio/workpresence/utils"
	"github.com/zeebo/errs"
	"golang.org/x/image/draw"
)

type Config struct {
	Command     string `default:"spectacle -m -b -n -o {{.Output}} -d 0 -p" help:"command to run to get a png stored in {{.Output}}"`
	JPEGQuality int    `default:"1" help:"jpeg quality, 1 to 100, higher is better"`
	MaxWidth    int    `default:"200" help:"max width for screenshot"`
	Greyscale   bool   `default:"false" help:"if true, convert image to greyscale"`
}

type ScreenshotSource struct {
	cfg     Config
	cmdTmpl *template.Template
}

func NewSource(cfg Config) (*ScreenshotSource, error) {
	tmpl, err := template.New("command").Parse(cfg.Command)
	if err != nil {
		return nil, err
	}
	return &ScreenshotSource{
		cfg:     cfg,
		cmdTmpl: tmpl,
	}, nil
}

func (s *ScreenshotSource) Screenshot(ctx context.Context) (*utils.SerializedImage, error) {
	fh, err := os.CreateTemp("", "screenshot-*.png")
	if err != nil {
		return nil, errs.Wrap(err)
	}
	name := fh.Name()
	defer func() {
		err = errs.Combine(err, os.Remove(name))
	}()
	err = fh.Close()
	if err != nil {
		return nil, errs.Wrap(err)
	}

	var cmdBuf bytes.Buffer
	if err = s.cmdTmpl.Execute(&cmdBuf, struct {
		Output string
	}{
		Output: name,
	}); err != nil {
		return nil, errs.Wrap(err)
	}
	command := cmdBuf.String()

	out, err := exec.CommandContext(ctx, "/bin/sh", "-c", command).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("process error: %q\n%w\n%q", string(command), err, string(out))
	}

	fh, err = os.Open(name)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	defer func() {
		err = errs.Combine(err, errs.Wrap(fh.Close()))
	}()

	pixels, err := png.Decode(fh)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	if s.cfg.Greyscale {
		region := pixels.Bounds().Max
		toGrey := image.NewRGBA(image.Rect(0, 0, region.X, region.Y))
		for y := 0; y < region.Y; y++ {
			for x := 0; x < region.X; x++ {
				toGrey.Set(x, y, color.GrayModel.Convert(pixels.At(x, y)))
			}
		}
		pixels = toGrey
	}

	if pixels.Bounds().Max.X > s.cfg.MaxWidth {
		newX := s.cfg.MaxWidth
		newY := pixels.Bounds().Max.Y * s.cfg.MaxWidth / pixels.Bounds().Max.X

		resized := image.NewRGBA(image.Rect(0, 0, newX, newY))
		draw.NearestNeighbor.Scale(resized, resized.Rect, pixels, pixels.Bounds(), draw.Over, nil)
		pixels = resized
	}

	var outBuf bytes.Buffer
	err = jpeg.Encode(&outBuf, pixels, &jpeg.Options{Quality: s.cfg.JPEGQuality})
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return &utils.SerializedImage{
		Data:      outBuf.Bytes(),
		Extension: ".jpg",
		MIMEType:  "image/jpeg",
	}, nil
}
