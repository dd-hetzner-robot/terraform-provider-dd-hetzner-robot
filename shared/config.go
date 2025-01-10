package shared

import "net/http"

type ProviderConfig struct {
	Username string
	Password string
	URL      string
}

type HetznerRobotClient struct {
	config ProviderConfig
	client *http.Client
}

func NewHetznerRobotClient(config ProviderConfig) *HetznerRobotClient {
	return &HetznerRobotClient{
		config: config,
		client: &http.Client{},
	}
}

type VSwitchServer struct {
	ServerIP      string `json:"server_ip"`
	ServerIPv6Net string `json:"server_ipv6_net"`
	ServerNumber  int    `json:"server_number"`
	Status        string `json:"status"`
}

type VSwitchSubnet struct {
	IP      string `json:"ip"`
	Mask    int    `json:"mask"`
	Gateway string `json:"gateway"`
}

type VSwitchCloudNetwork struct {
	ID      int    `json:"id"`
	IP      string `json:"ip"`
	Mask    int    `json:"mask"`
	Gateway string `json:"gateway"`
}

type VSwitch struct {
	ID            int                   `json:"id"`
	Name          string                `json:"name"`
	VLAN          int                   `json:"vlan"`
	Cancelled     bool                  `json:"cancelled"`
	Servers       []VSwitchServer       `json:"servers,omitempty"`
	Subnets       []VSwitchSubnet       `json:"subnets,omitempty"`
	CloudNetworks []VSwitchCloudNetwork `json:"cloud_networks,omitempty"`
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func NewNotFoundError(message string) error {
	return &NotFoundError{Message: message}
}
