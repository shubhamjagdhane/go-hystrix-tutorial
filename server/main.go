package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

func init() {
	hystrix.ConfigureCommand("api", hystrix.CommandConfig{
		Timeout:                300, // in seconds
		RequestVolumeThreshold: 1,
		ErrorPercentThreshold:  100,
		SleepWindow:            15000,
	})
	// for hystrix dashboard
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go http.ListenAndServe(":8002", hystrixStreamHandler)
}

func main() {
	http.HandleFunc("/api", api)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func api(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()
	responseChan := make(chan string, 1)

	errChan := hystrix.Go("api", func() error {
		res, err := http.Get("http://localhost:8081")
		if err != nil {
			fmt.Fprintf(w, "Error while hitting the client: %v", err)
			return err
		}
		defer res.Body.Close()

		bb, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Fprintf(w, "Error while reading response body: %v", err)
			return err
		}

		responseChan <- fmt.Sprintf("%s\n", bb)
		return nil
	}, func(err error) error {
		fmt.Println(err)
		return err
	})

	select {
	case <-ctx.Done():
		fmt.Fprintf(w, "Timeout\n")
	case <-errChan:
		fmt.Fprintf(w, "client maybe down, please try again later\n")
	case res := <-responseChan:
		fmt.Fprintf(w, res)
	}
}
