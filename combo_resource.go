package main

import(
	"errors"
)

var comboResource *ComboResource

type ComboResource struct {
	PrivateToken    string
	//PreGitSshUrl    string
	PerGitHttpUrl   string
	Registry        string
	RegistryCommon  string
	MingbaiHost     string
	MingbaiRegistry string
}

func (ComboResource) GetInstance(comboResourceStr, registryCommon,urlGitlab,privateTokenGitlab *string) (*ComboResource,error) {

		registry := ""
		if StringPointCheck(registryCommon) {
			registry = *registryCommon
		}
		switch *comboResourceStr {
		case "p2shop":
			comboResource = &ComboResource{
				PrivateToken:    p2shopToken,
				PerGitHttpUrl:   "https://gitlab.p2shop.cn:8443",
				Registry:        "registry.p2shop.com.cn",
				RegistryCommon:  registry,
				MingbaiHost:     "https://gateway.p2shop.com.cn",
				MingbaiRegistry: "registry.p2shop.com.cn",
			}
		case "srx":
			comboResource = &ComboResource{
				PrivateToken:    srxToken,
				PerGitHttpUrl:   "https://gitlab.srxcloud.com",
				Registry:        "registry.p2shop.com.cn",
				RegistryCommon:  registry,
				MingbaiHost:     "https://gateway.p2shop.com.cn",
				MingbaiRegistry: "swr.cn-north-1.myhuaweicloud.com/srx-cloud",
			}
		case "srx-p2shop":
			comboResource = &ComboResource{
				PrivateToken:    srxToken,
				PerGitHttpUrl:   "https://gitlab.srxcloud.com",
				Registry:        "registry.p2shop.com.cn",
				RegistryCommon:  registry,
				MingbaiHost:     "https://gateway.p2shop.com.cn",
				MingbaiRegistry: "registry.p2shop.com.cn",
			}
			fallthrough
		default:
		}
		if StringPointCheck(urlGitlab) {
			comboResource.PerGitHttpUrl = *urlGitlab
		}
		if StringPointCheck(privateTokenGitlab) {
			comboResource.PrivateToken = *privateTokenGitlab
		}
		if len(comboResource.PerGitHttpUrl) ==0 || len(comboResource.PrivateToken)==0{
			return nil,errors.New("Please check if the following environment variables are configured: REGISTRY_P2SHOP_PWD, GITLAB_P2SHOP_PRIVATETOKEN, GITLAB_SRX_PRIVATETOKEN")
		}
	return comboResource,nil
}
