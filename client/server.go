package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

func (c *HetznerRobotClient) FetchServerByIDWithContext(ctx context.Context, id int) (Server, error) {
	path := fmt.Sprintf("/server/%d", id)

	resp, err := c.DoRequest("GET", path, nil, "")
	if err != nil {
		return Server{}, fmt.Errorf("error fetching server: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return Server{}, fmt.Errorf("server with ID %d not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return Server{}, fmt.Errorf("error fetching server: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Server Server `json:"server"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Server{}, fmt.Errorf("error decoding server response: %w", err)
	}

	return result.Server, nil
}

func (c *HetznerRobotClient) FetchServersByIDs(ids []int) ([]Server, error) {
	var (
		servers []Server
		mu      sync.Mutex
		wg      sync.WaitGroup
		errs    []error
	)

	sem := make(chan struct{}, 10) // Ограничение на количество горутин
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, id := range ids {
		wg.Add(1)
		go func(serverID int) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			server, err := c.FetchServerByIDWithContext(ctx, serverID)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			servers = append(servers, server)
			mu.Unlock()
		}(id)
	}

	wg.Wait()

	if len(errs) > 0 {
		firstErrors := errs
		if len(errs) > 5 {
			firstErrors = errs[:5]
		}
		return nil, fmt.Errorf("errors occurred: %v (and %d more)", firstErrors, len(errs)-len(firstErrors))
	}

	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Number < servers[j].Number
	})

	return servers, nil
}
