package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	director := func(r *http.Request) {
		r.URL.Host = "storage.googleapis.com"
		r.URL.Scheme = "https"

		r.Header.Del("Cookie")
		r.Header.Del("Accept-Encoding")
	}

	modifyResponse := func(w *http.Response) error {
		w.Header.Del("x-goog-generation")
		w.Header.Del("x-goog-metageneration")
		w.Header.Del("x-goog-stored-content-encoding")
		w.Header.Del("x-goog-stored-content-length")
		w.Header.Del("x-goog-hash")
		w.Header.Del("x-goog-storage-class")
		w.Header.Del("x-goog-meta-goog-reserved-file-mtime")
		w.Header.Del("x-guploader-uploadid")
		w.Header.Del("Alt-Svc")
		w.Header.Del("Server")
		w.Header.Del("Age")

		return nil
	}

	rev := &httputil.ReverseProxy{
		Director:       director,
		Transport:      transport,
		ModifyResponse: modifyResponse,
	}

	srv := http.Server{
		Addr:    ":8080",
		Handler: rev,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "could not start http server: %s\n", err)
			os.Exit(1)
		}
	}()

	go func() {
		// health check
		http.ListenAndServe(":18080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "could not shutdown http server: %s\n", err)
	}
}
