package browser

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wirepair/gcd"
	"gitlab.com/browserker/browserk"
)

// GCDBrowserPool manages a pool of browsers via a leaser interface
type GCDBrowserPool struct {
	profileDir       string
	maxBrowsers      int
	acquiredBrowsers int32
	acquireErrors    int32
	browsers         chan *gcd.Gcd
	browserTimeout   time.Duration
	closing          int32
	display          string
	leaser           LeaserService
	startCount       int32
	logger           zerolog.Logger
}

// NewGCDBrowserPool number of pools, and a leaser that we can use
func NewGCDBrowserPool(maxBrowsers int, leaser LeaserService) *GCDBrowserPool {
	b := &GCDBrowserPool{}
	b.maxBrowsers = maxBrowsers
	b.browserTimeout = time.Second * 45
	b.leaser = leaser
	b.browsers = make(chan *gcd.Gcd, b.maxBrowsers)
	return b
}

// UseDisplay (to be called before Init()) tells chrome to start using an Xvfb display
func (b *GCDBrowserPool) UseDisplay(display string) {
	b.display = fmt.Sprintf("DISPLAY=%s", display)
}

// Init starts the browser/Browser pool
func (b *GCDBrowserPool) Init() error {
	return b.Start()
}

// SetAPITimeout tells gcd how long to wait for a response from the browser for all API calls
func (b *GCDBrowserPool) SetAPITimeout(duration time.Duration) {
	b.browserTimeout = duration
}

// Start the browser with a random profile directory and create Browsers
func (b *GCDBrowserPool) Start() error {
	// allow 3 seconds per Browser
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(b.maxBrowsers*3))
	defer cancel()

	// clean up just in case we are restarting
	if _, err := b.leaser.Cleanup(); err != nil {
		panic(fmt.Sprintf("failed to clean up browsers %s", err))
	}

	log.Info().Int("browsers", b.maxBrowsers).Msg("creating browsers")
	b.browsers = make(chan *gcd.Gcd, b.maxBrowsers)

	atomic.AddInt32(&b.startCount, 1)
	currentCount := atomic.LoadInt32(&b.startCount)
	// always have 2 browsers ready
	for i := 0; i < b.maxBrowsers; i++ {
		b.returnBrowser(timeoutCtx, "", currentCount) // passing an empty port will just create a new one for us
		log.Info().Int("i", i).Msg("browser created")
	}

	time.Sleep(time.Second * 2) // give time for browser to settle
	return nil
}

// Acquire a Browser, unless context expired. If expired, increment our Browser error count
// which is used to restart the entire browser process after a max limit on errors
// is reached
func (b *GCDBrowserPool) Acquire(ctx context.Context) *gcd.Gcd {

	select {
	case br := <-b.browsers:
		if br != nil {
			atomic.AddInt32(&b.acquiredBrowsers, 1)
		}
		return br
	case <-ctx.Done():
		log.Warn().Err(ctx.Err()).Msg("failed to acquire Browser from pool")
		atomic.AddInt32(&b.acquireErrors, 1)
		b.shouldRestart()
		return nil
	}
}

// Closing a channel that may be being read will cause a panic, which is fine because
// then we just restart anyways
func (b *GCDBrowserPool) shouldRestart() {
	acquired := atomic.LoadInt32(&b.acquiredBrowsers)
	errored := atomic.LoadInt32(&b.acquireErrors)
	count, _ := b.leaser.Count()
	log.Warn().Int32("acquired", acquired).Int32("errored", errored).Str("leaser_count", count).Msg("force restarting due to failure to acquire browsers")
	// flag as shutting down and clear out errors
	atomic.StoreInt32(&b.closing, 1)
	atomic.StoreInt32(&b.acquiredBrowsers, 0)
	atomic.StoreInt32(&b.acquireErrors, 0)
	// empty pool
	for {
		select {
		case <-b.browsers:
			log.Info().Msg("emptying browser")
		default:
			goto EMPTY
		}
	}
EMPTY:
	time.Sleep(1 * time.Second)
	log.Info().Msg("calling restart")
	if err := b.Start(); err != nil {
		panic("restarting due to failure to restart browsers process")
	}
	atomic.StoreInt32(&b.closing, 0)
}

// Return a browser
func (b *GCDBrowserPool) returnBrowser(ctx context.Context, port string, startCount int32) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	doneCh := make(chan struct{})

	go b.closeAndCreateBrowser(port, doneCh, startCount)

	select {
	case <-timeoutCtx.Done():
		log.Error().Msg("failed to closeAndCreateBrowser in time")
		return
	case <-ctx.Done():
		log.Error().Msg("ctx done")
		return
	case <-doneCh:
		return
	}
}

// closeAndCreateBrowser takes an optional Browser to close, and creates a new one, closing doneCh
// to signal it completed (although it may be a nil browser if error occurred).
func (b *GCDBrowserPool) closeAndCreateBrowser(port string, doneCh chan struct{}, startCount int32) {
	if port != "" {
		if err := b.leaser.Return(port); err != nil {
			log.Error().Err(err).Msg("failed to return browser")
		}
		atomic.AddInt32(&b.acquiredBrowsers, -1)
	}

	// if we've restarted and this browser was still leased, we don't want to create another one
	currentCount := atomic.LoadInt32(&b.startCount)
	if currentCount != startCount {
		close(doneCh)
		return
	}

	newBr := gcd.NewChromeDebugger()
	port, err := b.leaser.Acquire()
	if err != nil {
		log.Warn().Err(err).Msg("unable to acquire new browser")
		b.browsers <- nil
		close(doneCh)
		return
	}

	if err := newBr.ConnectToInstance("localhost", string(port)); err != nil {
		log.Warn().Err(err).Msg("failed to connect to instance")
		newBr = nil
	}
	newBr.SetTimeout(b.browserTimeout)
	b.browsers <- newBr
	close(doneCh)
}

// Take a browser
func (b *GCDBrowserPool) Take(ctx *browserk.Context) (browserk.Browser, string, error) {
	var br *gcd.Gcd

	if atomic.LoadInt32(&b.closing) == 1 {
		return nil, "", ErrBrowserClosing
	}
	// if nil, do not return browser
	if br = b.Acquire(ctx.Ctx); br == nil {
		return nil, "", errors.New("browser acquisition failed during Take")
	}

	log.Info().Int32("acquired", atomic.LoadInt32(&b.acquiredBrowsers)).Int32("errors", atomic.LoadInt32(&b.acquireErrors)).Msg("acquired browser")
	t, err := br.GetFirstTab()
	if err != nil {
		b.Return(ctx.Ctx, br.Port())
		return nil, "", fmt.Errorf("failed to aquire valid tab from browser")
	}
	gtab := NewTab(ctx, br, t)
	gtab.t.SetApiTimeout(b.browserTimeout) // default of 2 min is too long
	return gtab, br.Port(), nil
}

// Return a browser for destruction
func (b *GCDBrowserPool) Return(ctx context.Context, browserPort string) {
	startCount := atomic.LoadInt32(&b.startCount) // track if we've restarted so we can throw away bad browsers
	log.Info().Msg("closing browser")
	b.returnBrowser(ctx, browserPort, startCount)
	return
}

// Close all browsers and return. TODO: make this not terrible.
func (b *GCDBrowserPool) Close(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&b.closing, 0, 1) {
		return nil
	}

	for {
		br := b.Acquire(ctx)
		if br != nil {
			if err := b.leaser.Return(br.Port()); err != nil {
				log.Error().Err(err).Msg("failed to return browser")
			}
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if len(b.browsers) == 0 {
			return nil
		}
	}
}

// Leased returns the number of open/active browsers
func (b *GCDBrowserPool) Leased() int {
	return int(atomic.LoadInt32(&b.acquiredBrowsers))
}

// Shutdown kills browsers
func (b *GCDBrowserPool) Shutdown() error {
	atomic.CompareAndSwapInt32(&b.closing, 0, 1)
	_, err := b.leaser.Cleanup()
	return err
}
