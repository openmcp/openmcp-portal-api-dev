package handler

import (
	"github.com/jinzhu/configor"
)

var portalConfig = struct {
	Portal struct {
		OpenmcpURL string
		Port       string
		Kubeconfig string
	}
}{}

func InitPortalConfig() string {
	configor.Load(&portalConfig, "portalConfig.yml")
	return portalConfig.Portal.OpenmcpURL + ":" + portalConfig.Portal.Port
}

var openmcpURL = InitPortalConfig()
var kubeConfigFile = portalConfig.Portal.Kubeconfig
