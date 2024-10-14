// internal/middleware/csp.go

package middleware

import (
	"net/http"
)

type CSPOptions struct {
	DefaultSrc              string
	ScriptSrc               string
	ObjectSrc               string
	StyleSrc                string
	ImgSrc                  string
	ConnectSrc              string
	FontSrc                 string
	FrameSrc                string
	MediaSrc                string
	ReportURI               string
	UpgradeInsecureRequests bool
}

func CSP(options CSPOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			csp := ""
			if options.DefaultSrc != "" {
				csp += "default-src " + options.DefaultSrc + "; "
			}
			if options.ScriptSrc != "" {
				csp += "script-src " + options.ScriptSrc + "; "
			}
			if options.ObjectSrc != "" {
				csp += "object-src " + options.ObjectSrc + "; "
			}
			if options.StyleSrc != "" {
				csp += "style-src " + options.StyleSrc + "; "
			}
			if options.ImgSrc != "" {
				csp += "img-src " + options.ImgSrc + "; "
			}
			if options.ConnectSrc != "" {
				csp += "connect-src " + options.ConnectSrc + "; "
			}
			if options.FontSrc != "" {
				csp += "font-src " + options.FontSrc + "; "
			}
			if options.FrameSrc != "" {
				csp += "frame-src " + options.FrameSrc + "; "
			}
			if options.MediaSrc != "" {
				csp += "media-src " + options.MediaSrc + "; "
			}
			if options.UpgradeInsecureRequests {
				csp += "upgrade-insecure-requests; "
			}
			if options.ReportURI != "" {
				csp += "report-uri " + options.ReportURI + "; "
			}

			w.Header().Set("Content-Security-Policy", csp)
			next.ServeHTTP(w, r)
		})
	}
}
