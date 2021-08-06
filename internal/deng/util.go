package deng

import (
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	KubernetesProvider = "kubernetes"
)

var jsonMarshaller = &jsonpb.Marshaler{
	OrigName:     true,
	EnumsAsInts:  true,
	EmitDefaults: false,
}

func (x *Environment) GetOptionsAsJson() (string, error) {
	if k := x.GetKubernetes(); k != nil {
		return jsonMarshaller.MarshalToString(k)
	}
	if a := x.GetAws(); a != nil {
		return jsonMarshaller.MarshalToString(a)
	}
	return "{}", nil
}

func (x *Environment) ReadOptionsFromJson(provider string, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if provider == "" {
		provider = x.Provider
	}
	switch provider {
	case KubernetesProvider:
		opts := &KubernetesQualifier{}
		if err := json.Unmarshal(data, opts); err != nil {
			return err
		}
		x.Qualifier = &Environment_Kubernetes{
			Kubernetes: opts,
		}
		return nil
	}
	// TODO implement more
	return nil
}

func IntOrPercentFromInt(val uint32) *IntOrPercent {
	return &IntOrPercent{ValPct: &IntOrPercent_Value{
		Value: val,
	}}
}

func IntOrPercentFromPercent(val float64) *IntOrPercent {
	return &IntOrPercent{ValPct: &IntOrPercent_Percent{
		Percent: val,
	}}
}

func ConvertIntStr(x *IntOrString) *intstr.IntOrString {
	if x == nil {
		return nil
	}
	var i intstr.IntOrString
	v := x.GetSValue()
	if v != "" {
		i = intstr.FromString(v)
	} else {
		i = intstr.FromInt(int(x.GetIValue()))
	}
	return &i
}

func IntOrStringFromString(val string) IntOrString {
	return IntOrString{Value: &IntOrString_SValue{SValue: val}}
}

func IntOrStringFromInt(val int32) IntOrString {
	return IntOrString{Value: &IntOrString_IValue{IValue: val}}
}

var priorityStatus = []Status{
	Status_FAILED_CLEANING,
	Status_FAILED,
	Status_ABORTED,
	Status_PAUSED,
	Status_PENDING,
	Status_RESOLVED,
	Status_QUEUED,
	Status_SUCCEEDED_CLEANING,
	Status_SUCCEEDED,
	Status_NOT_STARTED,
}

var statusConversion = map[AtomicStatus]Status{
	AtomicStatus_NotStarted: Status_NOT_STARTED,
	AtomicStatus_Initiated:  Status_PENDING,
	AtomicStatus_Ready:      Status_PENDING,
	AtomicStatus_Rolling:    Status_PENDING,
	AtomicStatus_Paused:     Status_PAUSED,
	AtomicStatus_Success:    Status_SUCCEEDED,
	AtomicStatus_Failure:    Status_FAILED,
	AtomicStatus_Aborted:    Status_ABORTED,
}

func CombineStatus(s1, s2 Status) Status {
	for _, s := range priorityStatus {
		if s1 == s || s2 == s {
			return s
		}
	}
	return Status_FAILED
}

func AtomicToDeploymentStatus(s AtomicStatus) Status {
	return statusConversion[s]
}

func (s Status) IsFinal() bool {
	switch s {
	case Status_FAILED, Status_ABORTED, Status_FAILED_CLEANING, Status_SUCCEEDED:
		return true
	}
	return false
}

func (x *Strategy) GetDescription() string {
	if x.GetCanary() != nil {
		return "canary"
	} else if x.GetBlueGreen() != nil {
		return "blue/green"
	} else if x.GetRolling() != nil {
		return "rolling update"
	} else if x.GetUpdate() == true {
		return "update"
	} else if x.GetRecreate() != nil {
		return "recreate"
	}
	return "other"
}
