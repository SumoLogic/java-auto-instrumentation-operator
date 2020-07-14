package controller

import (
	"github.com/SumoLogic/java-auto-instrumentation-operator/pkg/controller/javaautoinstrumentationoperator"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, javaautoinstrumentationoperator.Add)
}
