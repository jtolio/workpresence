package screenshots

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"text/template"

	"github.com/zeebo/errs"
)

type Config struct {
	Command     string `default:"spectacle -m -b -n -o {{.Output}} -d 0 -p" help:"command to run to get a png stored in {{.Output}}"`
	JPEGQuality int    `default:"1" help:"jpeg quality, 1 to 100, higher is better"`
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

func (s *ScreenshotSource) Screenshot(ctx context.Context) (img []byte, extension, mimetype string, err error) {
	fh, err := os.CreateTemp("", "screenshot-*.png")
	if err != nil {
		return nil, "", "", err
	}
	name := fh.Name()
	defer func() {
		err = errs.Combine(err, os.Remove(name))
	}()
	err = fh.Close()
	if err != nil {
		return nil, "", "", err
	}

	var cmdBuf bytes.Buffer
	if err = s.cmdTmpl.Execute(&cmdBuf, struct {
		Output string
	}{
		Output: name,
	}); err != nil {
		return nil, "", "", err
	}
	command := cmdBuf.String()

	out, err := exec.CommandContext(ctx, "/bin/sh", "-c", command).CombinedOutput()
	if err != nil {
		return nil, "", "", fmt.Errorf("process error: %q\n%w\n%q", string(command), err, string(out))
	}

	fh, err = os.Open(name)
	if err != nil {
		return nil, "", "", err
	}
	defer func() {
		err = errs.Combine(err, fh.Close())
	}()

	pixels, err := png.Decode(fh)
	if err != nil {
		return nil, "", "", err
	}

	var outBuf bytes.Buffer
	err = jpeg.Encode(&outBuf, pixels, &jpeg.Options{Quality: s.cfg.JPEGQuality})
	if err != nil {
		return nil, "", "", err
	}

	return outBuf.Bytes(), ".jpg", "image/jpeg", nil
}
