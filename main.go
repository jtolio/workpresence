package main

import (
	"context"

	"github.com/dsnet/try"
	"github.com/spf13/pflag"
	"gopkg.in/webhelp.v1/whfatal"
	"gopkg.in/webhelp.v1/whlog"
	"storj.io/private/cfgstruct"

	"github.com/jtolio/workpresence/screenshots"
	"github.com/jtolio/workpresence/server"
	"github.com/jtolio/workpresence/storage"
)

var cfg struct {
	Addr        string `default:"127.0.0.1:3333" help:"address to listen on"`
	Server      server.Config
	Screenshots screenshots.Config
	Storage     storage.Config
}

func init() { cfgstruct.Bind(pflag.CommandLine, &cfg) }

func main() {
	pflag.Parse()
	ctx := context.Background()

	source := try.E1(screenshots.NewSource(cfg.Screenshots))

	dest := try.E1(storage.NewImageDest(ctx, cfg.Storage))
	defer dest.Close()

	s := server.New(cfg.Server, source, dest)

	go func() {
		try.E(whlog.ListenAndServe(cfg.Addr,
			whlog.LogRequests(logDebug, whlog.LogResponses(logDebug,
				whfatal.Catch(tryShim(s))))))
	}()

	try.E(s.Run(ctx))
}
