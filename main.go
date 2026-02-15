package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		log.Fatalf("Cannot create state dir: %v", err)
	}

	var allocCancel context.CancelFunc

	if cdpURL != "" {
		log.Printf("Connecting to Chrome at %s", cdpURL)
		bridge.allocCtx, allocCancel = chromedp.NewRemoteAllocator(context.Background(), cdpURL)
	} else {
		if err := os.MkdirAll(profileDir, 0755); err != nil {
			log.Fatalf("Cannot create profile dir: %v", err)
		}
		log.Printf("Launching Chrome (profile: %s, headless: %v)", profileDir, headless)

		opts := []chromedp.ExecAllocatorOption{
			chromedp.UserDataDir(profileDir),
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,

			// Stealth: disable automation indicators
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.Flag("exclude-switches", "enable-automation"),
			chromedp.Flag("disable-infobars", true),
			chromedp.Flag("disable-background-networking", false),
			chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
			chromedp.Flag("disable-popup-blocking", true),
			chromedp.Flag("disable-default-apps", false),
			chromedp.Flag("no-first-run", true),
			// Suppress "didn't shut down correctly" restore bar
			chromedp.Flag("disable-session-crashed-bubble", true),
			chromedp.Flag("hide-crash-restore-bubble", true),

			chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"),
			chromedp.WindowSize(1440, 900),
		}

		if headless {
			opts = append(opts, chromedp.Headless)
		} else {
			opts = append(opts, chromedp.Flag("headless", false))
		}

		markCleanExit()
		bridge.allocCtx, allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	}
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(bridge.allocCtx)
	defer browserCancel()

	// Inject stealth scripts
	if err := chromedp.Run(browserCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(`
				Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
				if (!window.chrome) { window.chrome = {}; }
				if (!window.chrome.runtime) { window.chrome.runtime = {}; }
				const originalQuery = window.navigator.permissions.query;
				window.navigator.permissions.query = (parameters) => (
					parameters.name === 'notifications' ?
						Promise.resolve({ state: Notification.permission }) :
						originalQuery(parameters)
				);
				Object.defineProperty(navigator, 'plugins', {
					get: () => [1, 2, 3, 4, 5],
				});
				Object.defineProperty(navigator, 'languages', {
					get: () => ['en-GB', 'en-US', 'en'],
				});
			`).Do(ctx)
			return err
		}),
	); err != nil {
		log.Fatalf("Cannot start Chrome: %v", err)
	}

	bridge.browserCtx = browserCtx
	bridge.tabs = make(map[string]*TabEntry)
	bridge.snapshots = make(map[string]*refCache)

	// Register the initial tab
	initTargetID := chromedp.FromContext(browserCtx).Target.TargetID
	bridge.tabs[string(initTargetID)] = &TabEntry{ctx: browserCtx}
	log.Printf("Initial tab: %s", initTargetID)

	if !noRestore {
		restoreState()
	}

	// Background tab cleanup
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	go bridge.cleanStaleTabs(cleanupCtx, 30*actionTimeout)

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /tabs", handleTabs)
	mux.HandleFunc("GET /snapshot", handleSnapshot)
	mux.HandleFunc("GET /screenshot", handleScreenshot)
	mux.HandleFunc("GET /text", handleText)
	mux.HandleFunc("POST /navigate", handleNavigate)
	mux.HandleFunc("POST /action", handleAction)
	mux.HandleFunc("POST /evaluate", handleEvaluate)
	mux.HandleFunc("POST /tab", handleTab)

	srv := &http.Server{Addr: ":" + port, Handler: corsMiddleware(authMiddleware(mux))}

	// Graceful shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down, saving state...")
		cleanupCancel()
		saveState()
		markCleanExit()
		shutdownCtx, shutdownDone := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownDone()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	log.Printf("🦀 PINCH! PINCH! — Pinchtab running on http://localhost:%s", port)
	log.Printf("   CDP target: %s", cdpURL)
	if token != "" {
		log.Println("   Auth: Bearer token required")
	} else {
		log.Println("   Auth: none (set BRIDGE_TOKEN to enable)")
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
