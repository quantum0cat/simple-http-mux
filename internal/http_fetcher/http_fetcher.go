package http_fetcher

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"simple-http-mux/internal/models"
	"simple-http-mux/pkg/errgroup"
	"simple-http-mux/pkg/utils"
	"time"
)

func FetchUrls(
	ctx context.Context,
	urls []string,
	maxWorkers int,
	fetchTimeout time.Duration,
	requestTimeout time.Duration,
) ([]models.Response, error) {

	//validate
	if len(urls) == 0 {
		return nil, errors.New("unable to fetch URLs, the input URL list is empty")
	}
	//remove url duplicates, we don't need to do extra job
	urls = utils.RemoveDuplicates(urls)

	if maxWorkers < 1 {
		maxWorkers = 1
	}
	if maxWorkers > len(urls) {
		maxWorkers = len(urls)
	}

	urlCh := make(chan string)
	respCh := make(chan models.Response)

	//generate input for workers
	go func(urls []string, urlCh chan string) {
		for _, url := range urls {
			urlCh <- url
		}
	}(urls, urlCh)

	var cancel context.CancelFunc
	if fetchTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, fetchTimeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)

	//init workers
	for wid := 0; wid < maxWorkers; wid++ {
		errGroup.Go(func() error {
			err := fetchUrlWorker(ctx, urlCh, respCh, requestTimeout)
			return err
		})
	}

	//collect results
	var responses []models.Response
	errGroup.Go(func() error {
		for i := 0; i < len(urls); i++ {
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
		return nil, err
	}
	log.Printf("Fetch finished")
	return responses, nil

}

func fetchUrl(ctx context.Context, client *http.Client, url string) (*models.Response, error) {
	log.Printf("Fetching %s...", url)
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
func fetchUrlWorker(
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
				resp, err := fetchUrl(ctx, &client, url)
				if err != nil {
					return errors.New(fmt.Sprintf("failed to fetch '%s': %s", url, err.Error()))
				}
				resultCh <- *resp
			}

		}
	}

}
