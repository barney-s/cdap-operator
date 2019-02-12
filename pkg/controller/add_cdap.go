package controller

import (
	"io.cdap/cdap-operator/pkg/controller/cdap"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, cdap.Add)
}
