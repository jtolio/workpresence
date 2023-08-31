package storage

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/jtolio/workpresence/utils"
	"github.com/zeebo/errs"
	"storj.io/uplink"
)

type Config struct {
	UplinkAccess     string        `help:"storj uplink access grant"`
	ListingAccessKey string        `help:"storj listing access key"`
	ListingSecretKey string        `help:"storj listing secret key"`
	Bucket           string        `help:"storj bucket"`
	PathPrefix       string        `help:"storj path prefix"`
	Expiration       time.Duration `default:"1h" help:"when screenshots expire. 0 means no expiration."`
}

type ImageDest struct {
	cfg  Config
	proj *uplink.Project
}

func NewImageDest(ctx context.Context, cfg Config) (*ImageDest, error) {
	access, err := uplink.ParseAccess(cfg.UplinkAccess)
	if err != nil {
		return nil, err
	}
	proj, err := uplink.OpenProject(ctx, access)
	if err != nil {
		return nil, err
	}
	d := &ImageDest{
		cfg:  cfg,
		proj: proj,
	}
	err = d.init(ctx)
	return d, err
}

func (d *ImageDest) upload(ctx context.Context, path string, data []byte, mimeType string, expiration time.Time) error {
	w, err := d.proj.UploadObject(ctx, d.cfg.Bucket, d.cfg.PathPrefix+path,
		&uplink.UploadOptions{Expires: expiration})
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return errs.Combine(err, w.Abort())
	}
	if mimeType != "" {
		err = w.SetCustomMetadata(ctx, uplink.CustomMetadata{
			"Content-Type": mimeType,
		})
	}
	return w.Commit()
}

func (d *ImageDest) init(ctx context.Context) error {
	_, err := d.proj.StatObject(ctx, d.cfg.Bucket, d.cfg.PathPrefix+"index.html")
	if err == nil || !errors.Is(err, uplink.ErrObjectNotFound) {
		return err
	}
	var out bytes.Buffer
	err = indexHTML.Execute(&out, struct {
		Config Config
	}{
		Config: d.cfg,
	})
	if err != nil {
		return err
	}
	return d.upload(ctx, "index.html", out.Bytes(), "text/html", time.Time{})
}

func (d *ImageDest) Store(ctx context.Context, ts time.Time, img *utils.SerializedImage) error {
	pathname := ts.UTC().Format("2006/01/02/15/04-05" + img.Extension)
	var expiration time.Time
	if d.cfg.Expiration > 0 {
		expiration = time.Now().Add(d.cfg.Expiration)
	}
	return d.upload(ctx, pathname, img.Data, img.MIMEType, expiration)
}

func (d *ImageDest) Close() error {
	return d.proj.Close()
}
