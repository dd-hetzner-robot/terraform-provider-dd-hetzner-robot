package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (c *HetznerRobotClient) FetchVSwitchByIDWithContext(ctx context.Context, id string) (*VSwitch, error) {
	resp, err := c.DoRequest("GET", fmt.Sprintf("/vswitch/%s", id), nil, "")
	if err != nil {
		return nil, fmt.Errorf("error fetching VSwitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("VSwitch with ID %s not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error fetching VSwitch: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	var vswitch VSwitch
	if err := json.NewDecoder(resp.Body).Decode(&vswitch); err != nil {
		return nil, fmt.Errorf("error decoding VSwitch response: %w", err)
	}

	return &vswitch, nil
}

func (c *HetznerRobotClient) FetchVSwitchesByIDs(ids []string) ([]VSwitch, error) {
	var (
		vswitches []VSwitch
		mu        sync.Mutex
		wg        sync.WaitGroup
		errs      []error
	)

	sem := make(chan struct{}, 10) // Ограничение параллелизма
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, id := range ids {
		wg.Add(1)
		go func(vswitchID string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			vswitch, err := c.FetchVSwitchByIDWithContext(ctx, vswitchID)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			vswitches = append(vswitches, *vswitch)
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

	sort.Slice(vswitches, func(i, j int) bool {
		return vswitches[i].ID < vswitches[j].ID
	})

	return vswitches, nil
}

func (c *HetznerRobotClient) CreateVSwitch(ctx context.Context, name string, vlan int) (*VSwitch, error) {
	data := url.Values{}
	data.Set("name", name)
	data.Set("vlan", strconv.Itoa(vlan))

	resp, err := c.DoRequest("POST", "/vswitch", strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return nil, fmt.Errorf("error creating VSwitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error creating VSwitch: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	var vswitch VSwitch
	if err := json.NewDecoder(resp.Body).Decode(&vswitch); err != nil {
		return nil, fmt.Errorf("error decoding VSwitch response: %w", err)
	}

	return &vswitch, nil
}

func (c *HetznerRobotClient) UpdateVSwitch(ctx context.Context, id, name string, vlan int) error {
	data := url.Values{}
	data.Set("name", name)
	data.Set("vlan", strconv.Itoa(vlan))

	resp, err := c.DoRequest("PUT", fmt.Sprintf("/vswitch/%s", id), strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return fmt.Errorf("error updating VSwitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error updating VSwitch: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *HetznerRobotClient) DeleteVSwitch(ctx context.Context, id string, cancellationDate string) error {
	data := url.Values{}
	data.Set("cancellation_date", cancellationDate)

	resp, err := c.DoRequest("DELETE", fmt.Sprintf("/vswitch/%s", id), strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return fmt.Errorf("error deleting VSwitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting VSwitch: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *HetznerRobotClient) AddVSwitchServers(ctx context.Context, id string, servers []VSwitchServer) error {
	data := url.Values{}
	for _, server := range servers {
		data.Add("server[]", strconv.Itoa(server.ServerNumber)) // Изменено на "server[]"
	}

	fmt.Printf("Adding servers to vSwitch %s with data: %v\n", id, data.Encode()) // Отладочный вывод

	resp, err := c.DoRequest("POST", fmt.Sprintf("/vswitch/%s/server", id), strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return fmt.Errorf("error adding servers to VSwitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error adding servers to VSwitch: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *HetznerRobotClient) RemoveVSwitchServers(ctx context.Context, id string, servers []VSwitchServer) error {
	data := url.Values{}
	for _, server := range servers {
		data.Add("server_number", strconv.Itoa(server.ServerNumber))
	}

	resp, err := c.DoRequest("DELETE", fmt.Sprintf("/vswitch/%s/server", id), strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return fmt.Errorf("error removing servers from VSwitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error removing servers from VSwitch: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *HetznerRobotClient) SetVSwitchCancellation(ctx context.Context, id, cancellationDate string) error {
	data := url.Values{}
	data.Set("cancellation_date", cancellationDate)

	resp, err := c.DoRequest("POST", fmt.Sprintf("/vswitch/%s/cancel", id), strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return fmt.Errorf("error setting cancellation date: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
