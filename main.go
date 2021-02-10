package main

import (
	"context"
	"net/http"
	"os"

	"github.com/factorysh/eat-my-beats/eat"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func main() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	f, err := os.OpenFile("./out.log", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	defer f.Close()
	b := eat.New(f)
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go b.Start(ctx)
	http.Handle("/beats/", b.Mux)
	listen := os.Getenv("LISTEN")
	if listen == "" {
		listen = ":9200"
	}
	log.Info().Str("listen", listen).Msg("")
	err = http.ListenAndServe(listen, nil)
	if err != nil {
		log.Error().Err(err).Msg("")
	}
}
