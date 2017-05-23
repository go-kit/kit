package consul

import (
	"strconv"

	consul "github.com/hashicorp/consul/api"
)

// Client TODO
type Client interface {
	Service(service, tag string, passingOnly bool, waitIndex uint64) ([]string, uint64, error)
}

// NewClient TODO
func NewClient(c *consul.Client) Client {
	return realClient{c}
}

type realClient struct{ *consul.Client }

func (c realClient) Service(service, tag string, passingOnly bool, waitIndex uint64) ([]string, uint64, error) {
	entries, meta, err := c.Client.Health().Service(service, tag, passingOnly, &consul.QueryOptions{WaitIndex: waitIndex})
	if err != nil {
		return []string{}, 0, err
	}
	results := make([]string, len(entries))
	for i := 0; i < len(entries); i++ {
		results[i] = entries[i].Node.Address + ":" + strconv.Itoa(entries[i].Service.Port)
	}
	return results, meta.LastIndex, nil
}
