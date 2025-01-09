package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *HetznerRobotClient) CreateVSwitch(name string, vlan int) (*VSwitch, error) {
	form := fmt.Sprintf("name=%s&vlan=%d", name, vlan)
	reqBody := bytes.NewBufferString(form)
	resp, err := c.DoRequest("POST", "/vswitch", reqBody, "application/x-www-form-urlencoded")
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

func (c *HetznerRobotClient) GetVSwitchByID(id string) (*VSwitch, error) {
	resp, err := c.DoRequest("GET", fmt.Sprintf("/vswitch/%s", id), nil, "")
	if err != nil {
		return nil, fmt.Errorf("error getting VSwitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("VSwitch with ID %s not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error getting VSwitch: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	var vswitch VSwitch
	if err := json.NewDecoder(resp.Body).Decode(&vswitch); err != nil {
		return nil, fmt.Errorf("error decoding VSwitch response: %w", err)
	}

	return &vswitch, nil
}

func (c *HetznerRobotClient) UpdateVSwitch(id, name string, vlan int) error {
	form := fmt.Sprintf("name=%s&vlan=%d", name, vlan)
	reqBody := bytes.NewBufferString(form)
	resp, err := c.DoRequest("PUT", fmt.Sprintf("/vswitch/%s", id), reqBody, "application/x-www-form-urlencoded")
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

func (c *HetznerRobotClient) DeleteVSwitch(id string) error {
	resp, err := c.DoRequest("DELETE", fmt.Sprintf("/vswitch/%s", id), nil, "")
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
