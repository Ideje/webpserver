package cache

import (
	"github.com/chai2010/webp"
	"github.com/davecgh/go-spew/spew"
	"image"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"webpserver/app"
	"webpserver/config"
)

type Resource struct {
	Host string
	Path string
}

// MakeResource create resource
func MakeResource(host, path string) *Resource {
	return &Resource{
		Host: host,
		Path: path,
	}
}

// Src return url w/o scheme
func (rs *Resource) Src() string {
	return "//" + rs.Host + strings.Replace("/"+rs.Path, "//", "/", -1)
}

// URL return url w/ scheme
func (rs *Resource) URL() string {
	return "http:" + rs.Src()
}

// SrcPath return base path like data/<prefix>/domain.com/a.jpg
func (rs *Resource) SrcPath() (path string) {
	path = config.Cfg.ImageSrcDir + rs.Path
	return
}

func (rs *Resource) SrcOrigPath() (path string) {
	ext := filepath.Ext(rs.Path)
	//parts := strings.Split(rs.Path[:len(rs.Path)-len(ext)], "-")
	path = config.Cfg.ImageSrcDir + rs.Path + ext
	return
}

// CreateDstFile return created os.File w/ path like data/<prefix>/domain.com/a.jpg.<suffix>.<hash>
func (rs *Resource) CreateDstFile() (file *os.File, err error) {
	dir, fileName := filepath.Split(rs.DstPath())
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return
	}
	file, err = ioutil.TempFile(dir, fileName+".*.tmp")
	app.LogError(err)
	return
}

// DstPath return base path like data/<prefix>/a.jpg.webp
func (rs *Resource) DstPath() (path string) {
	path = config.Cfg.ImageDstDir + rs.Path + ".webp"
	return
}

func (rs *Resource) IsActual(srcPath string) (bool, int64, int64, error) {
	var err error

	srcStat, err := os.Stat(srcPath)
	if err != nil {
		app.Log.Warn("File not found", "path", srcPath)
		return false, 0, 0, err
	}

	dstStat, err := os.Stat(rs.DstPath())
	if err != nil {
		return false, srcStat.Size(), 0, nil
	}

	return dstStat.ModTime().After(srcStat.ModTime()), srcStat.Size(), dstStat.Size(), nil
}

func ImageLoad(filename string) (m image.Image, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	m, _, err = image.Decode(f)
	return
}

func ImageSaveWebp(m image.Image, file *os.File, opts *webp.Options) (err error) {
	defer file.Close()

	err = webp.Encode(file, m, opts)
	if err != nil {
		app.Log.Error("err", "err", err)
		return
	}

	err = file.Sync()
	app.LogError(err)

	return
}

func WebpCreate(src string, dst *os.File) (err error) {
	img, err := ImageLoad(src)
	if err != nil {
		app.Log.Warn("ImageLoad", "err", err, "src", src)
		return err
	}

	ext := filepath.Ext(src)

	fi, err := os.Stat(src)
	if err != nil {
		// Could not obtain stat, handle error
		app.LogError(err)
		return err
	}
	app.Log.Debug("Src file", "path", src, "size", fi.Size(), "ext", ext)

	switch ext {
	case ".jpg", ".jpeg":
		webpOpts := &webp.Options{
			Lossless: false,
			Quality:  85,
		}
		err := ImageSaveWebp(img, dst, webpOpts)
		app.LogError(err)
	case ".png", ".gif":
		webpOpts := &webp.Options{
			Lossless: true,
		}
		err := ImageSaveWebp(img, dst, webpOpts)
		app.LogError(err)
	case ".webp":
		source, err := os.Open(src)
		if err != nil {
			app.LogError(err)
			return err
		}
		defer source.Close()

		_, err = io.Copy(dst, source)
		app.LogError(err)
	default:
		return nil
	}
	return err
}

func (rs *Resource) Process(srcPath string) (err error) {
	app.Log.Debug("Process", "src", srcPath)
	dstFile, err := rs.CreateDstFile()
	if err != nil {
		return err
	}

	defer func() {
		os.Remove(dstFile.Name())
	}()

	err = WebpCreate(srcPath, dstFile)
	if err != nil {
		spew.Dump(err)
		return err
	}

	err = os.Rename(dstFile.Name(), rs.DstPath())
	app.LogError(err)

	return nil
}
