package main

// Provides output for use as an Ansible inventory plugin

import (
	"encoding/json"

	"github.com/jimyhuang/linode"
)

type inventory struct {
	Meta  map[string]map[string]map[string]string `json:"_meta"`
	Hosts []string                                `json:"hosts"`
}

var inv = inventory{}

func newInventory(nodes map[int]*linodeWithIPs) {
	hostvars := make(map[string]map[string]string, 0)
	if inv.Meta == nil {
		meta := make(map[string]map[string]map[string]string)
		meta["hostvars"] = hostvars
		inv = inventory{Meta: meta, Hosts: make([]string, 0)}
	}
	for _, n := range nodes {
		inv.Hosts = append(inv.Hosts, n.node.Label)
		publicIP, privateIP, rdns := publicPrivateIP(n.ips)
		inv.Meta["hostvars"][n.node.Label] = map[string]string{
			"ansible_ssh_host":   publicIP,
			"host_label":         n.node.Label,
			"host_display_group": n.node.DisplayGroup,
			"host_private_ip":    privateIP,
			"host_public_ip":     publicIP,
			"host_rdns":          rdns,
		}
	}
}

func toJSON() ([]byte, error) {
	return json.MarshalIndent(inv, " ", "  ")
}

func publicPrivateIP(ips []linode.LinodeIP) (string, string, string) {
	var pub, prv, rdns string
	for _, ip := range ips {
		if ip.IsPublic() {
			pub = ip.IP
		} else {
			prv = ip.IP
		}
		rdns = ip.RDNS
		if pub != "" && prv != "" {
			break
		}
	}
	return pub, prv, rdns
}

func invMerge(x1, x2 interface{}) (interface{}, error) {
	data1, err := json.Marshal(x1)
	if err != nil {
		return nil, err
	}
	data2, err := json.Marshal(x2)
	if err != nil {
		return nil, err
	}
	var j1 interface{}
	err = json.Unmarshal(data1, &j1)
	if err != nil {
		return nil, err
	}
	var j2 interface{}
	err = json.Unmarshal(data2, &j2)
	if err != nil {
		return nil, err
	}
	return invMerge1(j1, j2), nil
}
func invMerge1(x1, x2 interface{}) interface{} {
	switch x1 := x1.(type) {
	case map[string]interface{}:
		x2, ok := x2.(map[string]interface{})
		if !ok {
			return x1
		}
		for k, v2 := range x2 {
			if v1, ok := x1[k]; ok {
				x1[k] = invMerge1(v1, v2)
			} else {
				x1[k] = v2
			}
		}
	case nil:
		// merge(nil, map[string]interface{...}) -> map[string]interface{...}
		x2, ok := x2.(map[string]interface{})
		if ok {
			return x2
		}
	}
	return x1
}
