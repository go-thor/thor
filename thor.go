package thor

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-thor/thor/build"
	"github.com/go-thor/thor/logger"
	"github.com/go-thor/thor/server"
	"golang.org/x/sync/errgroup"
)

type (
	Application interface {
		ID() string
		Name() string
		Version() string
		Namespace() string
		Run() error
	}

	// application app interface
	application struct {
		opts *Options
		quit chan os.Signal
	}

	serverErr struct {
		server server.Server
		err    error
	}
)

// New a server bootstrap
func New(options ...Option) Application {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	if opts.startupTimeout == 0 {
		opts.startupTimeout = 1000
	}
	if opts.shutdownTimeout == 0 {
		opts.shutdownTimeout = 5000
	}
	if opts.log == nil {
		opts.log = logger.Nop
	}

	app := &application{
		quit: make(chan os.Signal),
		opts: opts,
	}

	signal.Notify(app.quit, syscall.SIGINT, syscall.SIGTERM)

	return app
}

func (app *application) ID() string {
	return build.ID
}

func (app *application) Name() string {
	return build.Name
}

func (app *application) Version() string {
	return build.Version
}

func (app *application) Namespace() string {
	return build.Namespace
}

func (app *application) Run() error {
	if err := app.serve(); err != nil {
		return err
	}

	defer close(app.quit)
	<-app.quit

	if err := app.shutdown(); err != nil {
		return err
	}

	return nil
}

func (app *application) serve() error {
	app.opts.log.Info("serve start...")

	g := errgroup.Group{}
	for _, b := range app.opts.servers {
		b := b
		g.Go(func() error {
			var (
				hook        server.Hook
				ctx, cancel = context.WithTimeout(context.Background(), time.Duration(app.opts.startupTimeout)*time.Millisecond)
			)
			defer cancel()

			if v, ok := b.(server.Hook); ok {
				hook = v
			}

			if hook != nil {
				if err := hook.BeforeServe(ctx); err != nil {
					return err
				}
			}

			if err := app.startServer(ctx, b); err != nil {
				return err
			}

			if hook != nil {
				if err := hook.AfterServe(ctx); err != nil {
					return err
				}
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		app.opts.log.Errorf("serve error: %v", err)
	} else {
		app.opts.log.Infof("serve done...")
	}

	return err
}

func (app *application) startServer(ctx context.Context, b server.Server) error {
	var (
		serv    server.Server
		serveCh = make(chan serverErr)
	)

	if v, ok := b.(server.Server); ok {
		serv = v
	}

	go func() {
		defer close(serveCh)

		if serv != nil {
			app.opts.log.Infof("serve %s", serv.Name())
			serveCh <- serverErr{
				server: b,
				err:    serv.Serve(ctx),
			}
		}
	}()

	select {
	case ch := <-serveCh:
		if ch.err != nil {
			app.opts.log.Errorf("serve %s error: %v", ch.server.Name(), ch.err)
		}

		return ch.err
	case <-ctx.Done():
		if ctx.Err() != context.DeadlineExceeded {
			return ctx.Err()
		}
	}

	return nil
}

func (app *application) shutdown() error {
	app.opts.log.Info("shutdown start...")

	g := errgroup.Group{}
	for _, b := range app.opts.servers {
		b := b
		g.Go(func() error {
			var (
				hook        server.Hook
				ctx, cancel = context.WithTimeout(context.Background(), time.Duration(app.opts.shutdownTimeout)*time.Millisecond)
			)
			defer cancel()

			if v, ok := b.(server.Hook); ok {
				hook = v
			}

			if hook != nil {
				if err := hook.BeforeShutdown(ctx); err != nil {
					return err
				}
			}

			if err := app.shutdownServer(ctx, b); err != nil {
				return err
			}

			if hook != nil {
				if err := hook.AfterShutdown(ctx); err != nil {
					return err
				}
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		app.opts.log.Errorf("shutdown error: %v", err)
	} else {
		app.opts.log.Infof("shutdown done...")
	}

	return err
}

func (app *application) shutdownServer(ctx context.Context, b server.Server) error {
	var (
		serv       server.Server
		shutdownCh = make(chan serverErr)
	)

	if v, ok := b.(server.Server); ok {
		serv = v
	}

	go func() {
		defer close(shutdownCh)

		if serv != nil {
			app.opts.log.Infof("shutdown %s", serv.Name())
			shutdownCh <- serverErr{
				server: b,
				err:    serv.Shutdown(ctx),
			}
		}
	}()

	select {
	case ch := <-shutdownCh:
		if ch.err != nil {
			app.opts.log.Errorf("shutdown %s error: %v", ch.server.Name(), ch.err)
		}
		return ch.err
	case <-ctx.Done():
		return ctx.Err()
	}
}
