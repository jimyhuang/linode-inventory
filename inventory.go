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
