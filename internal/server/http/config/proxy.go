package config

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrIndexArrByString    = errors.New("cannot index array with string")
	ErrIndexArrOutOfBounds = errors.New("out of bounds array index")
)

type Proxy struct {
	Host url.URL `yaml:"host"`
}

func NewProxy(imap map[interface{}]interface{}) (*Proxy, error) {
	var proxy Proxy

	if x, ok := imap["host"]; ok {
		u, err := url.Parse(x.(string))
		if err != nil {
			return nil, err
		}

		proxy.Host = *u
	}

	return &proxy, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	redirect(&p.Host, r)

	response, err := http.DefaultClient.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer response.Body.Close()

	var header = w.Header()
	for k, v := range response.Header {
		for _, h := range v {
			header.Add(k, h)
		}
	}

	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}

func redirect(target *url.URL, req *http.Request) {
	var (
		targetQuery = target.RawQuery

		singleJoiningSlash = func(a, b string) string {
			aslash := strings.HasSuffix(a, "/")
			bslash := strings.HasPrefix(b, "/")
			switch {
			case aslash && bslash:
				return a + b[1:]
			case !aslash && !bslash:
				return a + "/" + b
			}
			return a + b
		}

		joinURLPath = func(a, b *url.URL) (path, rawpath string) {
			if a.RawPath == "" && b.RawPath == "" {
				return singleJoiningSlash(a.Path, b.Path), ""
			}
			// Same as singleJoiningSlash, but uses EscapedPath to determine
			// whether a slash should be added
			apath := a.EscapedPath()
			bpath := b.EscapedPath()

			aslash := strings.HasSuffix(apath, "/")
			bslash := strings.HasPrefix(bpath, "/")

			switch {
			case aslash && bslash:
				return a.Path + b.Path[1:], apath + bpath[1:]
			case !aslash && !bslash:
				return a.Path + "/" + b.Path, apath + "/" + bpath
			}
			return a.Path + b.Path, apath + bpath
		}
	)

	req.RequestURI = ""
	req.Host = target.Host

	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
}
