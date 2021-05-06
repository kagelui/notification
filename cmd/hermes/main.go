package main

import (
	"context"
	"github.com/kagelui/notification/internal/models/bmodels"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kagelui/notification/internal/pkg/envvar"
	"github.com/kagelui/notification/internal/pkg/loglib"
	"github.com/kagelui/notification/internal/service/messages"
	_ "github.com/lib/pq"
)

const goRoutineCap = 10

func main() {
	lg := loglib.DefaultLogger()
	ctx := loglib.SetLogger(context.Background(), lg)

	lg.InfoF("starting retrying failed callback...")

	var e envVar

	if err := envvar.Read(&e); err != nil {
		lg.ErrorF(err.Error())
		os.Exit(1)
	}

	db, err := sqlx.Connect("postgres", e.DBAddr)
	if err != nil {
		lg.ErrorF(err.Error())
		os.Exit(2)
	}

	messageSlice, err := messages.RetrieveAllRetryMessages(ctx, db)
	if err != nil {
		lg.ErrorF(err.Error())
		os.Exit(3)
	}
	lg.InfoF("retrieved %v messages to retry", len(messageSlice))

	if err = messages.MarkMessagesPending(ctx, db, messageSlice); err != nil {
		lg.ErrorF(err.Error())
		os.Exit(4)
	}

	messageChannel := make(chan *bmodels.Message, goRoutineCap)
	var wg sync.WaitGroup
	for i := 1; i <= goRoutineCap; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			lg.InfoF("Starting runner %d", num)

			httpClient := http.DefaultClient
			httpClient.Timeout = e.ClientTimeout
			client := messages.CallbackClient{
				Client: httpClient,
			}

			for message := range messageChannel {
				lg.InfoF("starting callback %v", message.ID)
				if err := client.DoCallback(ctx, db, message); err != nil {
					lg.ErrorF("error doing callback for %v: %v", message.ID, err.Error())
				}
				lg.InfoF("sent callback %v", message.ID)
			}
			lg.InfoF("End runner %d", num)
		}(i)
	}
	for _, m := range messageSlice {
		messageChannel <- m
	}
	close(messageChannel)
	wg.Wait()

	lg.InfoF("end retry callbacks")
}

type envVar struct {
	DBAddr        string        `env:"DATABASE_URL"`
	ClientTimeout time.Duration `env:"CLIENT_TIMEOUT"`
}
