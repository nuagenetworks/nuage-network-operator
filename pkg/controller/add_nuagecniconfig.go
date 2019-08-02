package controller

import (
	"github.com/nuagenetworks/nuage-network-operator/pkg/controller/nuagecniconfig"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, nuagecniconfig.Add)
}
