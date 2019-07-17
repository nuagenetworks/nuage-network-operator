package controller

import (
	"github.com/nuagenetworks/nuage-network-operator/pkg/controller/network"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, network.Add)
}
