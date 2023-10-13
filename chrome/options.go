package chrome

import (
	"net/url"
	"time"

	"github.com/chromedp/chromedp"
)

// LocalChromeOptions Referenced from: https://github.com/puppeteer/puppeteer/issues/3938#issuecomment-475986157
var LocalChromeOptions = append(chromedp.DefaultExecAllocatorOptions[:],
	chromedp.DisableGPU,
	chromedp.NoSandbox,
	chromedp.Headless,
	chromedp.Flag("use-gl", "swiftshader"),
	chromedp.Flag("no-zygote", true),
	chromedp.Flag("disable-setuid-sandbox", true),
	chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"),
	chromedp.Flag("font-antialiasing", "1"),
	chromedp.Flag("virtual-time-budget", "10000"),
	chromedp.Flag("ignore-certificate-errors", true),
)

type ScreenshotOptions struct {
	URL     *url.URL
	Width   int64
	Height  int64
	Delay   time.Duration
	EndTime time.Time
	Full    bool
}

type NewTabOptions struct {
	URL     *url.URL
	Delay   time.Duration
	EndTime time.Time
}
