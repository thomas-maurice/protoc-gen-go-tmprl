package generator

import (
	"fmt"

	temporalv1 "git.maurice.fr/thomas/protoc-gen-go-tmprl/gen/temporal/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

type MethodType string

const (
	MethodTypeWorkflow = MethodType("WORKFLOW")
	MethodTypeActivity = MethodType("ACTIVITY")
	MethodTypeNone     = MethodType("NONE")
	MethodTypeInvalid  = MethodType("INVALID")
)

func getMethodType(m *protogen.Method) (MethodType, error) {
	wf, _ := proto.GetExtension(m.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions)
	act, _ := proto.GetExtension(m.Desc.Options(), temporalv1.E_Activity).(*temporalv1.ActivityOptions)

	if wf == nil && act == nil {
		return MethodTypeNone, nil
	}

	if wf != nil && act != nil {
		return MethodTypeInvalid, fmt.Errorf("invalid method %s, cannot be both an activity and a workflow", m.Desc.Name())
	}

	if act != nil {
		return MethodTypeActivity, nil
	}

	return MethodTypeWorkflow, nil
}

func getMethodRegisteredName(m *protogen.Method) (string, error) {
	wf, _ := proto.GetExtension(m.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions)
	act, _ := proto.GetExtension(m.Desc.Options(), temporalv1.E_Activity).(*temporalv1.ActivityOptions)

	if wf == nil && act == nil {
		return "", nil
	}

	if wf != nil && act != nil {
		return "", fmt.Errorf("invalid method %s, cannot be both an activity and a workflow", m.Desc.Name())
	}

	if act != nil {
		if act.Name != "" {
			return act.Name, nil
		} else {
			return string(m.Desc.FullName()), nil
		}
	}

	if wf.Name != "" {
		return wf.Name, nil
	} else {
		return string(m.Desc.FullName()), nil
	}
}
