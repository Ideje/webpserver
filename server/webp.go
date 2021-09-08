package server

import (
	"fmt"
	"github.com/chai2010/webp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"webpserver/app"
	"webpserver/cache"
	"webpserver/config"
)

var (
	webpOpts = &webp.Options{
		Lossless: true,
		Quality:  90,
		Exact:    true,
	}
)

type webpHandler struct{}

func WebpServer() http.Handler {
	return &webpHandler{}
}

func (f *webpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ext := filepath.Ext(r.URL.Path)

	app.Log.Debug("WebP handler", "path", r.URL.Path, "ext", ext)

	parts := strings.Split(r.URL.Path[:len(r.URL.Path)-len(ext)], "/")
	rs := cache.MakeResource("localhost", strings.Join(append([]string{""}, parts[1:]...), "/"))

	if ext != ".webp" || filepath.Ext(rs.SrcPath()) == "" {
		http.ServeFile(w, r, config.Cfg.ImageSrcDir+"/"+r.URL.Path)
		return
	}

	isActual, srcSize, dstSize, err := rs.IsActual(rs.SrcPath())
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("File not found: %s", r.URL.Path), 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	if !isActual {
		if err := rs.Process(rs.SrcPath()); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if config.Cfg.HeaderCacheInfo {
			w.Header().Add("cache", "miss")
			w.Header().Add("cache-content-length-orig", fmt.Sprintf("%d", srcSize))

			dstStat, err := os.Stat(rs.DstPath())
			if err != nil {
				dstSize = 0
			} else {
				dstSize := dstStat.Size()
				w.Header().Add("cache-compression", fmt.Sprintf("%.2f", float64(dstSize)/float64(srcSize)))
			}
		}
	} else {
		if config.Cfg.HeaderCacheInfo {
			w.Header().Add("cache", "hit")
			w.Header().Add("cache-content-length-orig", fmt.Sprintf("%d", srcSize))
			w.Header().Add("cache-compression", fmt.Sprintf("%.2f", float64(dstSize)/float64(srcSize)))
		}
	}

	http.ServeFile(w, r, rs.DstPath())
}
