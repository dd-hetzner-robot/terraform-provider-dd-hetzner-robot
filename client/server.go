package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

func (c *HetznerRobotClient) FetchServerByID(id int) (Server, error) {
	resp, err := c.DoRequest("GET", fmt.Sprintf("/server/%d", id), nil, "")
	if err != nil {
		return Server{}, fmt.Errorf("error fetching server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Server{}, NewNotFoundError(fmt.Sprintf("server with ID %d not found", id))
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

	for _, id := range ids {
		wg.Add(1)
		go func(serverID int) {
			defer wg.Done()

			server, err := c.FetchServerByID(serverID)
			if err != nil {
				if _, ok := err.(*NotFoundError); !ok {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
				return
			}

			mu.Lock()
			servers = append(servers, server)
			mu.Unlock()
		}(id)
	}

	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("errors occurred: %v", errs)
	}

	return servers, nil
}
