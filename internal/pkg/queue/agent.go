package jobqueue2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"github.com/kagelui/notification/internal/pkg/loglib"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Publisher is the interface for pubishing a message to the job queue
type Publisher interface {
	Publish(ctx context.Context, job interface{}, b backoff.BackOff) error
	Stop()
}

// workerFunc is a named type that represents the worker function signature
type workerFunc func(ctx context.Context, b []byte) error

// Make sure agent implements Publisher
var _ Publisher = (*agent)(nil)

// agent is the default jobqueue agent which serves as a publisher and manages worker func
type agent struct {
	amqpURL     string
	queueName   string
	ctx         context.Context
	ctxCancel   context.CancelFunc
	logger      *loglib.Logger
	connBackOff backoff.BackOff
	conn        *amqp.Connection
	ch          *amqp.Channel
	mu          sync.Mutex
	wg          sync.WaitGroup
	// workerCancel is a context.CancelFunc to cancel to agent's worker goroutine
	workerCancel context.CancelFunc
	// workerFunc is the worker func executed to process messages from the queue
	workerFunc workerFunc
	// autoAck is a flag to indicate whether the consumer will auto-acknowledge messages
	autoAck bool
}

// NewPublisher connects to rabbitmq, returns a publisher where you can call `Publish`
func NewPublisher(ctx context.Context, queueName, amqpURL string, connBackOff backoff.BackOff) (Publisher, error) {
	agentCtx, agentCancel := context.WithCancel(ctx)
	logger := loglib.GetLogger(agentCtx).
		WithField("queue", queueName).
		WithField("pid", strconv.Itoa(os.Getpid()))
	a := &agent{
		amqpURL:     amqpURL,
		queueName:   queueName,
		ctx:         agentCtx,
		ctxCancel:   agentCancel,
		logger:      logger,
		connBackOff: connBackOff,
	}
	return a, a.connect()
}

// NewWorker connects to rabbitmq, and invokes callback function for every message received
// `autoAck` was formerly default `false`
//
// when `autoAck` is false, watch out for rabbitmq re-queuing messages upon errors (e.g. connection drop)
// https://www.rabbitmq.com/confirms.html#automatic-requeueing
func NewWorker(ctx context.Context, queueName, amqpURL string, connBackOff backoff.BackOff, autoAck bool, callback workerFunc) (context.CancelFunc, error) {
	agentCtx, agentCancel := context.WithCancel(ctx)
	logger := loglib.GetLogger(agentCtx).
		WithField("queue", queueName).
		WithField("pid", strconv.Itoa(os.Getpid()))
	a := &agent{
		amqpURL:     amqpURL,
		queueName:   queueName,
		ctx:         agentCtx,
		ctxCancel:   agentCancel,
		logger:      logger,
		connBackOff: connBackOff,

		// everything above here is identical to NewPublisher
		workerFunc: recoverWorkerFunc(logger, callback),
		autoAck:    autoAck,
	}
	return a.Stop, a.connect()
}

func (a *agent) Stop() {
	a.ctxCancel()
	a.wg.Wait()
}

// Publish marshals a job into JSON and delivers it to rabbitmq
func (a *agent) Publish(ctx context.Context, job interface{}, b backoff.BackOff) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return backoff.Retry(func() error {
		a.mu.Lock()
		defer a.mu.Unlock()
		return a.ch.Publish(
			"",          // exchange
			a.queueName, // routing key
			false,       // mandatory
			false,
			amqp.Publishing{
				MessageId:    uuid.New().String(),
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         payload,
			})
	}, backoff.WithContext(b, a.ctx))
}

// launches a goroutine that consumes messages from a queue
// and calls `a.worker` function for each
// goroutine ends when connection errors, or context is done
func (a *agent) runWorker() error {
	if a.workerFunc == nil {
		return nil
	}

	consumerUUID := uuid.New().String()
	logger := a.logger.WithField("consumer_uuid", consumerUUID)

	// Stop an agent's running worker (if any) before spawning a new worker goroutine.
	// This is to ensure that there's only one active/running worker for each agent.
	if a.workerCancel != nil {
		logger.InfoF("stopping running workers, if any")
		a.workerCancel()
	}
	workerCtx, workerCancel := context.WithCancel(a.ctx)

	var msgs <-chan amqp.Delivery
	var err error

	a.mu.Lock()
	// Set worker cancel func to agent for reference
	a.workerCancel = workerCancel

	if err = a.ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	); err == nil {
		msgs, err = a.ch.Consume(
			a.queueName,  // queue
			consumerUUID, // consumer
			a.autoAck,    // auto-ack

			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
	}
	a.mu.Unlock()

	if err == nil {
		logger.InfoF("started worker")
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			defer workerCancel()

			for {
				select {
				case <-workerCtx.Done(): // Worker context cancelled/done
					logger.InfoF("worker context ended")
					return
				case <-a.ctx.Done(): // Agent context cancelled/done
					logger.InfoF("agent context ended")
					return
				case d, ok := <-msgs:
					if !ok {
						logger.InfoF("messages ended; close connection to trigger reconnect logic")
						if err := a.conn.Close(); err != nil {
							logger.ErrorF("close connection failed: %q", err.Error())
						}
						return
					}
					err = a.workerFunc(a.ctx, d.Body)
					logger.InfoF("received %#v %s", err, string(d.Body))
					if err != nil {
						a.reportError(err)
					}
					if !a.autoAck {
						// only this message; not "all messages before this"
						if err := d.Ack(false); err != nil {
							logger.ErrorF("ack failed: %q", err.Error())
						}
					}
				}
			}
		}()
	}

	return err
}

// [repeatedly] tries to connect (with backoff) to rabbitmq
func (a *agent) connect() error {
	err := backoff.Retry(func() error {
		if a.amqpURL == "" {
			return fmt.Errorf("no mq URL")
		}
		parts := strings.Split(a.amqpURL, "@")
		hostport := parts[len(parts)-1]

		a.logger.InfoF("dialing %q", hostport)
		conn, err := amqp.Dial(a.amqpURL)
		if err != nil {
			return errors.Wrapf(err, "dialing %q", hostport)
		}
		// Set connection to agent for reference
		a.mu.Lock()
		a.conn = conn
		a.mu.Unlock()

		if err := a.openChannel(); err != nil {
			return errors.Wrapf(err, "openChannel")
		}

		if err := a.runWorker(); err != nil {
			return errors.Wrapf(err, "startWorkers")
		}

		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.waitChannel()
		}()
		a.logger.InfoF("connected %q", hostport)
		return nil
	}, backoff.WithContext(a.connBackOff, a.ctx))
	if err != nil {
		a.logger.ErrorF("connect failed: %q", err.Error())
		return err
	}
	return nil
}

// sets up our rabbitmq queue to use
func (a *agent) openChannel() error {
	ch, err := a.conn.Channel()
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.ch = ch
	a.mu.Unlock()

	_, err = ch.QueueDeclare(
		a.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	return err
}

// blocks until connection error, or context is done
func (a *agent) waitChannel() {
	defer a.conn.Close()

	for {
		select {
		case <-a.ctx.Done():
			return // job is done
		case b := <-a.conn.NotifyBlocked(make(chan amqp.Blocking)):
			a.logger.InfoF("conn.NotifyBlocked: %#v", b) // optional info only; resume infinite loop
		case err := <-a.conn.NotifyClose(make(chan *amqp.Error)):
			a.logger.InfoF("conn.NotifyClose: %v", err)
			a.wg.Add(1)
			go func() {
				defer a.wg.Done()
				// reconnect logic passes jobqueue2/integrationtest/retryconn.Makefile
				if err := a.connect(); err != nil {
					a.reportError(err)
				}
			}()
			return // we leave; the `connect` will do its own waitChannel
		}
	}
}

// recoverWorkerFunc adds a defer/recover mechanism on the provided worker func
func recoverWorkerFunc(logger *loglib.Logger, next workerFunc) workerFunc {
	return func(ctx context.Context, b []byte) (err error) {
		// defer recover function to ensure this service won't panic
		// if the panic really happens, return err and log the panic information
		defer func() {
			if r := recover(); r != nil {
				// TODO: if we can put into pkg/errors stack instead
				// of error message that'll be even better
				err = errors.Errorf("panic: recovered %s - %s", r, debug.Stack())
			}
		}()

		return next(ctx, b)
	}
}

// ReportError handles reporting of standard and Error errors
func (a *agent) reportError(err error) {
	if err == nil {
		return
	}

	logger := loglib.GetLogger(a.ctx)

	logger.ErrorF(err.Error())
	// additional error reporting if necessary
}
