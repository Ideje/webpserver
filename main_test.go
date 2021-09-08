package main

import (
	"github.com/bool64/httptestbench"
	"github.com/inconshreveable/log15"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"webpserver/app"
	"webpserver/config"
	"webpserver/server"
)

var (
	testServer           *httptest.Server
	benchmarkConcurrency int
	testSleep            time.Duration
)

func init() {
	benchmarkConcurrency = 64
	testSleep = time.Millisecond * 10
}

func checkResponseStatusCode(t *testing.T, expectedStatusCode int, res *http.Response) {
	if expectedStatusCode != res.StatusCode {
		t.Errorf("Expected response status code %d. Got %d\n", expectedStatusCode, res.StatusCode)
	}
}

func checkMinContentLength(t *testing.T, expectedMinContentLength int64, res *http.Response) {
	if expectedMinContentLength > res.ContentLength {
		t.Errorf("Expected minimal body length %d. Got %d\n", expectedMinContentLength, res.ContentLength)
	}
}

func RemoveDirFiles(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if strings.HasPrefix(name, ".") {
			continue
		}
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func TestMain(m *testing.M) {
	testServer = httptest.NewServer(server.Router())
	defer testServer.Close()

	app.Log.Info("test server", "base url", testServer.URL)

	RemoveDirFiles(config.Cfg.ImageDstDir)

	os.Exit(m.Run())
}

func testImageOk(t *testing.T, path string) {
	res, err := http.Get(testServer.URL + path)
	require.Nil(t, err)
	checkResponseStatusCode(t, http.StatusOK, res)
	checkMinContentLength(t, 10000, res)

	time.Sleep(testSleep)
}

func TestImageExistsJpeg(t *testing.T) {
	testImageOk(t, "/minions.jpg")
}

func TestImageExistsWebp(t *testing.T) {
	testImageOk(t, "/minions.webp")
}
func TestImageEncodeJpegWebp(t *testing.T) {
	testImageOk(t, "/minions.jpg.webp")
}
func TestImageCopyWebpWebp(t *testing.T) {
	testImageOk(t, "/minions.webp.webp")
}

func TestNotFound(t *testing.T) {
	for _, path := range [...]string{
		"/not-found",
		"/minions",
	} {
		res, err := http.Get(testServer.URL + path)
		require.Nil(t, err)
		checkResponseStatusCode(t, http.StatusNotFound, res)
	}
	time.Sleep(testSleep)
}

func Benchmark_jpg(b *testing.B) {
	benchmarkPath(b, "/minions.jpg")
}

func Benchmark_webp(b *testing.B) {
	benchmarkPath(b, "/minions.webp")
}

func Benchmark_jpg_webp(b *testing.B) {
	benchmarkPath(b, "/minions.jpg.webp")
}

func Benchmark_webp_webp(b *testing.B) {
	benchmarkPath(b, "/minions.webp.webp")
}

func benchmarkPath(b *testing.B, path string) {
	app.Log.SetHandler(log15.DiscardHandler())

	httptestbench.RoundTrip(b, benchmarkConcurrency,
		func(i int, req *fasthttp.Request) {
			req.SetRequestURI(testServer.URL + path)
		},
		func(i int, resp *fasthttp.Response) bool {
			return resp.StatusCode() == http.StatusOK
		},
	)
}
