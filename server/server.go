package server

import (
	"context"
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
	ScreenshotInterval       time.Duration `default:"10s" "the interval between screenshots"`
	ScreenshotIntervalJitter bool          `default:"false" "if true, will jitter screenshot interval"`
}

type ScreenshotSource interface {
	Screenshot(ctx context.Context) (data []byte, extension, mimetype string, err error)
}

type ScreenshotDest interface {
	Store(ctx context.Context, ts time.Time, img []byte, extension string) error
}

type Server struct {
	cfg    Config
	source ScreenshotSource
	dest   ScreenshotDest
	paused atomic.Bool
	latest atomic.Value

	http.Handler
}

func New(cfg Config, source ScreenshotSource, dest ScreenshotDest) *Server {
	s := &Server{
		cfg:    cfg,
		source: source,
		dest:   dest,
	}
	s.Handler = whmux.Dir{
		"":       whmux.Exact(http.HandlerFunc(s.pageLanding)),
		"pause":  whmux.ExactPath(http.HandlerFunc(s.pagePause)),
		"resume": whmux.ExactPath(http.HandlerFunc(s.pageResume)),
		"latest": whmux.Exact(http.HandlerFunc(s.pageLatest)),
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

type screenshot struct {
	data      []byte
	extension string
	mimetype  string
}

func (s *Server) Run(ctx context.Context) error {
	for time2.Sleep(ctx, s.interval()) {
		if s.paused.Load() {
			continue
		}
		ts := time.Now()
		img, extension, mimetype, err := s.source.Screenshot(ctx)
		if err != nil {
			slog.Error("failed to capture screenshot", err)
			continue
		}
		s.latest.Store(&screenshot{data: img, extension: extension, mimetype: mimetype})
		err = s.dest.Store(ctx, ts, img, extension)
		if err != nil {
			slog.Error("failed to store screenshot", err)
			continue
		}
	}
	return ctx.Err()
}

func (s *Server) pagePause(w http.ResponseWriter, r *http.Request) {
	s.paused.Store(true)
	whfatal.Redirect("/")
}

func (s *Server) pageResume(w http.ResponseWriter, r *http.Request) {
	s.paused.Store(false)
	whfatal.Redirect("/")
}

func (s *Server) pageLanding(w http.ResponseWriter, r *http.Request) {
	try.E1(w.Write([]byte(`<!doctype html><html><head>
	<meta http-equiv="refresh" content="5">
	</head><body>`)))
	if s.paused.Load() {
		try.E1(w.Write([]byte(`<p>Paused</p><p><a href="/resume">Resume</a></p>`)))
	} else {
		try.E1(w.Write([]byte(`<p>Running</p><p><a href="/pause">Pause</a></p>`)))
	}
	if s.latest.Load() != nil {
		try.E1(w.Write([]byte(`<img src="/latest" width=600>`)))
	}
	try.E1(w.Write([]byte(`</body></html>`)))
}

func (s *Server) pageLatest(w http.ResponseWriter, r *http.Request) {
	if latest, ok := s.latest.Load().(*screenshot); ok && latest != nil {
		w.Header().Set("Content-Type", latest.mimetype)
		try.E1(w.Write(latest.data))
	}
}
