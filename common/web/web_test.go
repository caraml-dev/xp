package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tu "github.com/caraml-dev/xp/common/testutils"
)

func startTestHTTPServer(mux *http.ServeMux, address string) *http.Server {
	srv := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	go func() {
		_ = srv.ListenAndServe()
	}()

	return srv
}

func TestFileHandler(t *testing.T) {
	mux := http.NewServeMux()

	filePath := "../testdata/config1.yaml"
	mux.Handle("/path", FileHandler(filePath, true))
	mux.Handle("/not-found", FileHandler(fmt.Sprintf("%d.file", time.Now().Unix()), false))

	srv := startTestHTTPServer(mux, ":9999")
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

	resp, httpErr := http.DefaultClient.Get("http://localhost:9999/path")
	require.NoError(t, httpErr)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
	require.Equal(t, "0", resp.Header.Get("Expires"))

	respBytes, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.NoError(t, err)

	data, _ := tu.ReadFile(filePath)
	require.Equal(t, data, respBytes)

	resp, httpErr = http.DefaultClient.Get("http://localhost:9999/not-found")
	require.NoError(t, httpErr)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	_ = resp.Body.Close()
}
