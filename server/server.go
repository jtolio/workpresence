package server

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/dsnet/try"
	"golang.org/x/exp/slog"
	"gopkg.in/webhelp.v1/whfatal"
	"gopkg.in/webhelp.v1/whmux"
	"storj.io/common/time2"

	"github.com/jtolio/workpresence/utils"
)

type Config struct {
	ScreenshotInterval       time.Duration `default:"45s" "the interval between screenshots"`
	ScreenshotIntervalJitter bool          `default:"false" "if true, will jitter screenshot interval"`
}

type ScreenshotSource interface {
	Screenshot(context.Context) (*utils.SerializedImage, error)
}

type ScreenshotDest interface {
	Store(context.Context, time.Time, *utils.SerializedImage) error
}

type Server struct {
	cfg    Config
	source ScreenshotSource
	dest   ScreenshotDest
	paused atomic.Bool
	latest atomic.Value
	cancel func()

	http.Handler
}

func New(cfg Config, source ScreenshotSource, dest ScreenshotDest) *Server {
	s := &Server{
		cfg:    cfg,
		source: source,
		dest:   dest,
	}
	s.Handler = whmux.Dir{
		"":            whmux.Exact(http.HandlerFunc(s.pageLanding)),
		"pause":       whmux.ExactPath(http.HandlerFunc(s.pagePause)),
		"resume":      whmux.ExactPath(http.HandlerFunc(s.pageResume)),
		"quit":        whmux.ExactPath(http.HandlerFunc(s.pageQuit)),
		"latest":      whmux.Exact(http.HandlerFunc(s.pageLatest)),
		"favicon.ico": whmux.Exact(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
	}
	return s
}

func (s *Server) interval() time.Duration {
	rv := s.cfg.ScreenshotInterval
	if s.cfg.ScreenshotIntervalJitter {
		rv = utils.NormJitter(rv) // should this be an exponential jitter?
	}
	if rv < time.Second {
		rv = time.Second
	}
	return rv
}

func (s *Server) takeScreenshot(ctx context.Context) {
	if s.paused.Load() {
		return
	}
	ts := time.Now()
	img, err := s.source.Screenshot(ctx)
	if err != nil {
		slog.Error("failed to capture screenshot", fmt.Errorf("%+w", err))
		return
	}
	s.latest.Store(img)
	err = s.dest.Store(ctx, ts, img)
	if err != nil {
		slog.Error("failed to store screenshot", err)
		return
	}
}

func (s *Server) Run(ctx context.Context) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.takeScreenshot(cancelCtx)
	for time2.Sleep(cancelCtx, s.interval()) {
		s.takeScreenshot(cancelCtx)
	}
	return ctx.Err()
}

func (s *Server) pagePause(w http.ResponseWriter, r *http.Request) {
	s.paused.Store(true)
	whfatal.Redirect("/")
}

func (s *Server) pageQuit(w http.ResponseWriter, r *http.Request) {
	cancel := s.cancel
	if cancel != nil {
		cancel()
	}
}

func (s *Server) pageResume(w http.ResponseWriter, r *http.Request) {
	s.paused.Store(false)
	s.takeScreenshot(r.Context())
	whfatal.Redirect("/")
}

func (s *Server) pageLanding(w http.ResponseWriter, r *http.Request) {
	try.E1(w.Write(indexHTMLHeader))
	if s.paused.Load() {
		try.E1(w.Write([]byte(`<p style="color:red;">Paused</p><p><a href="/resume">Resume</a></p>`)))
	} else {
		try.E1(w.Write([]byte(`<p>Running</p><p><a href="/pause">Pause</a></p>`)))
	}
	try.E1(w.Write(indexHTMLFooter))
}

func (s *Server) pageLatest(w http.ResponseWriter, r *http.Request) {
	if latest, ok := s.latest.Load().(*utils.SerializedImage); ok && latest != nil {
		w.Header().Set("Content-Type", latest.MIMEType)
		try.E1(w.Write(latest.Data))
	}
}
