package main

import "sync"

var comboResource *ComboResource
var comboResourceOnce sync.Once

type ComboResource struct {
	PrivateToken    string
	PreGitSshUrl    string
	PerGitHttpUrl   string
	Registry        string
	MingbaiHost     string
	MingbaiRegistry string
}

func (ComboResource) GetInstance(comboResourceStr *string) *ComboResource {

	comboResourceOnce.Do(func() {
		switch *comboResourceStr {
		case "msl":
			comboResource = &ComboResource{
				PrivateToken:    "bY2kmqs8x8N3wfQxgw6s",
				PreGitSshUrl:    "ssh://git@gitlab.p2shop.cn:822",
				PerGitHttpUrl:   "https://gitlab.p2shop.cn:8443",
				Registry:        "registry.elandsystems.cn",
				MingbaiHost:     "https://gateway.p2shop.com.cn",
				MingbaiRegistry: "registry.p2shop.com.cn",
			}
		case "srx":
			comboResource = &ComboResource{
				PrivateToken:    "SjPC8PnY6N8ntaxcUXFM",
				PreGitSshUrl:    "ssh://git@gitlab.srxcloud.com:622",
				PerGitHttpUrl:   "https://gitlab.srxcloud.com",
				Registry:        "registry.elandsystems.cn",
				MingbaiHost:     "https://gateway.p2shop.com.cn",
				MingbaiRegistry: "swr.cn-north-1.myhuaweicloud.com/srx-cloud",
			}
		case "srx-msl":
			comboResource = &ComboResource{
				PrivateToken:    "SjPC8PnY6N8ntaxcUXFM",
				PreGitSshUrl:    "ssh://git@gitlab.srxcloud.com:622",
				PerGitHttpUrl:   "https://gitlab.srxcloud.com",
				Registry:        "registry.elandsystems.cn",
				MingbaiHost:     "https://gateway.p2shop.com.cn",
				MingbaiRegistry: "registry.p2shop.com.cn",
			}
		}
	})
	return comboResource
}
