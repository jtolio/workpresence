package main

import (
	"fmt"
	"net/http"

	"github.com/dsnet/try"
	"golang.org/x/exp/slog"
	"gopkg.in/webhelp.v1/whfatal"
	"gopkg.in/webhelp.v1/whroute"
)

func tryShim(h http.Handler) http.Handler {
	return whroute.HandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer try.HandleF(&err, func() { whfatal.Error(err) })
		h.ServeHTTP(w, r)
	})
}

func logDebug(format string, arg ...interface{}) {
	slog.Debug(fmt.Sprintf(format, arg...))
}

func logInfo(format string, arg ...interface{}) {
	slog.Info(fmt.Sprintf(format, arg...))
}
