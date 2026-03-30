package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"

	"nexus/internal/gateway/router"
)

// WatchYAML debounces fs events and reloads routes into store.
func WatchYAML(ctx context.Context, path string, log *slog.Logger, set func(*router.Table) error) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := w.Add(path); err != nil {
		_ = w.Close()
		return err
	}

	var pending *time.Timer
	debounce := 300 * time.Millisecond

	go func() {
		defer func() { _ = w.Close() }()
		for {
			select {
			case <-ctx.Done():
				if pending != nil {
					pending.Stop()
				}
				return
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Chmod) == 0 {
					continue
				}
				if pending != nil {
					pending.Stop()
				}
				p := path
				pending = time.AfterFunc(debounce, func() {
					t, err := router.LoadYAML(p)
					if err != nil {
						if log != nil {
							log.Warn("routes reload failed", slog.String("path", p), slog.Any("err", err))
						}
						return
					}
					if err := set(t); err != nil && log != nil {
						log.Warn("routes swap failed", slog.Any("err", err))
					} else if log != nil {
						log.Info("routes reloaded", slog.String("path", p))
					}
				})
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				if log != nil {
					log.Warn("fsnotify error", slog.Any("err", err))
				}
			}
		}
	}()
	return nil
}
