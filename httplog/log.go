package log

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

var fgColorRed *color.Color
var bgWhiteFgColorRed *color.Color
var privateIPBlocks []*net.IPNet

func init() {
	fgColorRed = color.New(color.FgRed)
	bgWhiteFgColorRed = fgColorRed.Add(color.BgWhite)

	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func LogRequest(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{w, 200}
		h.ServeHTTP(rec, r)

		log.Print(strings.Join([]string{
			color.MagentaString(r.Host),
			bgWhiteFgColorRed.Sprint(getRemoteAddress(r)),
			getColoredStatusCode(rec.statusCode),
			r.Method,
			"\"" + color.CyanString(r.URL.String()) + "\"",
			"\"" + color.CyanString(r.Header.Get("User-Agent")) + "\"",
			time.Now().Sub(start).String(),
		}, " "))
	}

	return http.HandlerFunc(fn)
}

func getColoredStatusCode(code int) string {
	var colorFn func(string, ...interface{}) string
	if code < http.StatusMultipleChoices { // sucesses
		colorFn = color.HiGreenString
	} else if code < http.StatusInternalServerError { // client errors
		colorFn = color.HiBlueString
	} else { // server errors
		colorFn = color.HiRedString
	}
	return colorFn(strconv.Itoa(code))
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func getRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			// Returns first non-private IP address...
			parts[i] = strings.TrimSpace(p)
			if !isPrivateIP(net.ParseIP(parts[i])) {
				return parts[i]
			}
		}
	}
	return hdrRealIP
}
