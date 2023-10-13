package chrome

import (
	"context"
	"time"

	"github.com/4everland/screenshot/lib"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/inspector"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type Chrome struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

func NewLocalChrome(execPath, proxy string) *Chrome {
	if execPath != "" {
		LocalChromeOptions = append(LocalChromeOptions, chromedp.ExecPath(execPath))
	}

	if proxy != "" {
		LocalChromeOptions = append(LocalChromeOptions, chromedp.ProxyServer(proxy))
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), LocalChromeOptions...)
	return &Chrome{
		Ctx:    ctx,
		Cancel: cancel,
	}
}

func (c Chrome) Screenshot(parent context.Context, o ScreenshotOptions) (b []byte, err error) {
	timeoutCtx, cancelTimeoutCtx := context.WithTimeout(parent, time.Until(o.EndTime.Add(o.Delay)))
	defer cancelTimeoutCtx()

	ctx, cancel := chromedp.NewContext(timeoutCtx)
	defer cancel()

	/* --- prevent browser crashes from locking the context (prevents hanging) --- */
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if _, ok := ev.(*inspector.EventTargetCrashed); ok {
			cancel()
		}
	})

	chromedp.ListenTarget(timeoutCtx, func(ev interface{}) {
		if _, ok := ev.(*inspector.EventTargetCrashed); ok {
			cancelTimeoutCtx()
		}
	})
	/* --- */

	/* --- squash JavaScript dialog boxes such as alert(); --- */
	chromedp.ListenTarget(timeoutCtx, func(ev interface{}) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() {
				if err := chromedp.Run(timeoutCtx,
					page.HandleJavaScriptDialog(true),
				); err != nil {
					cancelTimeoutCtx()
				}
			}()
		}
	})
	/* --- */

	if err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.EmulateViewport(o.Width, o.Height),
		chromedp.Navigate(o.URL.String()),
		chromedp.Evaluate(`document.readyState === 'complete'`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, exc, err := runtime.Evaluate(`document.fonts.ready`).WithAwaitPromise(true).Do(ctx)
			if err != nil {
				return err
			}
			if exc != nil {
				return exc
			}
			return nil
		}),
		chromedp.Sleep(o.Delay),
		chromedp.Evaluate("document.querySelector('.osano-cm-dialog')?.remove()", nil),
		chromedp.Evaluate("document.querySelector('#onetrust-consent-sdk')?.remove()", nil),
		chromedp.Evaluate("document.querySelector('#hs-eu-cookie-confirmation')?.remove()", nil),
		chromedp.Evaluate("document.querySelector('#hubspot-messages-iframe-container')?.remove()", nil),

		chromedp.ActionFunc(func(ctx context.Context) error {
			if o.Full {
				return chromedp.FullScreenshot(&b, 100).Do(ctx)
			}

			return chromedp.CaptureScreenshot(&b).Do(ctx)
		}),
	}); err != nil {
		lib.Logger().Error("chrome screenshot err:"+err.Error(), lib.ChromeLog)
	}

	return b, err
}

func (c Chrome) RawHtml(parent context.Context, o NewTabOptions) (b string, err error) {
	timeoutCtx, cancel := context.WithTimeout(parent, time.Until(o.EndTime.Add(o.Delay)))
	defer cancel()

	ctx, cancel := chromedp.NewContext(timeoutCtx)
	defer cancel()

	if err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(o.URL.String()),
		chromedp.Sleep(o.Delay),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err := dom.GetOuterHTML().WithBackendNodeID(node.BackendNodeID).Do(ctx)
			if err == nil {
				b = html
			}
			return err
		}),
	}); err != nil {
		lib.Logger().Error("chrome catch html err:"+err.Error(), lib.ChromeLog)
	}

	return
}
