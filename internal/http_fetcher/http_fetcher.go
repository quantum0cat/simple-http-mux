/*
	The package implements an HttpFetcher, which is suited for getting responses for a list of URLs in a concurrent way.
*/
package http_fetcher

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"simple-http-mux/internal/models"
	"simple-http-mux/pkg/errgroup"
	"simple-http-mux/pkg/utils"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type HttpFetcher struct {
	rid            uint32        //request id, for debugging purposes
	urls           []string      //urls list to process
	maxWorkers     int           //max worker goroutines
	fetchTimeout   time.Duration //timeout to fetch all urls or cancel
	requestTimeout time.Duration //timeout for single request
}

func NewHttpFetcher(
	rid uint32,
	urls []string,
	maxWorkers int,
	fetchTimeout time.Duration,
	requestTimeout time.Duration,
) (*HttpFetcher, error) {

	//validate
	if len(urls) == 0 {
		return nil, fmt.Errorf("failed to construct an HttpFetcher, urls list is empty")
	}

	//remove url duplicates, we don't need to do extra job
	urls = utils.RemoveDuplicates(urls)

	if maxWorkers < 1 {
		maxWorkers = 1
	}
	if maxWorkers > len(urls) {
		maxWorkers = len(urls)
	}

	return &HttpFetcher{
			rid:            rid,
			urls:           urls,
			maxWorkers:     maxWorkers,
			fetchTimeout:   fetchTimeout,
			requestTimeout: requestTimeout,
		},
		nil
}

// Fetch
//fetches multiple urls concurrently, can be cancelled by ctx
func (h *HttpFetcher) Fetch(ctx context.Context) ([]models.Response, error) {
	log.Printf(utils.WithRid("Fetch started", h.rid))
	urlCh := make(chan string)
	respCh := make(chan models.Response)

	//generate input for workers
	go func(urls []string, urlCh chan string) {
		for _, url := range urls {
			urlCh <- url
		}
	}(h.urls, urlCh)

	var cancel context.CancelFunc
	if h.fetchTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, h.fetchTimeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)

	//init workers
	for wid := 0; wid < h.maxWorkers; wid++ {
		errGroup.Go(func() error {
			err := h.fetchUrlWorker(ctx, urlCh, respCh, h.requestTimeout)
			return err
		})
	}

	//collect results into resulting slice
	var responses []models.Response
	errGroup.Go(func() error {
		for i := 0; i < len(h.urls); i++ {
			select {
			case <-ctx.Done():
				return nil
			case res := <-respCh:
				responses = append(responses, res)
			}
		}
		cancel()
		return nil
	})

	//wait for all workers and results collector to finish
	err := errGroup.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf(utils.WithRid("Fetch was cancelled", h.rid))
			return nil, err
		}
		log.Printf(utils.WithRid("Fetch finished with error", h.rid))
		return nil, err
	}
	log.Printf(utils.WithRid("Fetch finished succesfully", h.rid))
	return responses, nil

}

//fetches single url with the given http client
func (h *HttpFetcher) fetchUrl(ctx context.Context, client *http.Client, url string) (*models.Response, error) {
	log.Printf("Fetching %s...", utils.WithRid(url, h.rid))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &models.Response{
			Url:      url,
			Response: string(body),
		},
		nil
}

//worker, that concurrently fetches any url, received from urlCh and moves the results into resultCh
//can be cancelled with the ctx
func (h *HttpFetcher) fetchUrlWorker(
	ctx context.Context,
	urlCh chan string,
	resultCh chan models.Response,
	requestTimeout time.Duration,
) error {
	//validate
	if requestTimeout < 0 {
		requestTimeout = 0
	}

	client := http.Client{
		Transport: http.DefaultTransport,
		Timeout:   requestTimeout,
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case url := <-urlCh:
			{
				resp, err := h.fetchUrl(ctx, &client, url)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return err
					}
					return errors.New(fmt.Sprintf("failed to fetch '%s': %s", url, err.Error()))
				}
				resultCh <- *resp
			}

		}
	}

}
