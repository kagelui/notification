package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/kagelui/notification/cmd/serverd/handler"
	"github.com/kagelui/notification/internal/pkg/envvar"
	"github.com/kagelui/notification/internal/pkg/server"
	"github.com/kagelui/notification/internal/service/messages"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("server started")

	var e envVar

	if err := envvar.Read(&e); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	db, err := sqlx.Connect("postgres", e.DBAddr)
	if err != nil {
		log.Println(err.Error())
		os.Exit(132)
	}

	modelStore := &messages.ModelStore{DB: db}

	r := mux.NewRouter()
	r.Handle("/callback", handler.WrapError(handler.StoreCallbackThenSend(modelStore, e.ClientTimeout))).Methods(http.MethodPost)

	server.New(":8080", r).Start()
}

type envVar struct {
	DBAddr string `env:"DATABASE_URL"`
	ClientTimeout time.Duration `env:"CLIENT_TIMEOUT"`
}
