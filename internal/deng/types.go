package deng

import (
	"time"
)

type AppSummary struct {
	Name           string
	Deployments    int32
	LastSuccessful time.Time
	LastFailure    time.Time
}

type AccountSummary struct {
	Provider string
	Name     string
}

type DeploymentConfiguration struct {
	Application string
	EnvironmentType string
	EnvironmentName string
	EnvironmentNamespace string
	ViaAccount string
	ViaProvider string
	Version []string
	Kustomize bool
	Local bool
	Strategy string
	StrategySteps []string
	Wait bool
}

type Environment struct {
	Provider string
	Account string
	Alias string
}

type PauseCondition struct {
	Reason    string
	StartTime time.Time
}

type RolloutCondition struct {
    // Type of deployment condition.
	Type string
	// Phase of the condition, one of True, False, Unknown.
	Status string
	// The last time this condition was updated.
	LastUpdateTime time.Time
	// Last time the condition transitioned from one status to another.
	LastTransitionTime time.Time
	// The reason for the condition's last transition.
	Reason string
	// A human-readable message indicating details about the transition.
	Message string
}

type KubernetesState struct {
	PauseConditions []*PauseCondition
	// CurrentPodHash the hash of the current pod template
	CurrentPodHash string
	// Total number of non-terminated pods targeted by this rollout (their labels match the selector).
	Replicas int32
	// Total number of non-terminated pods targeted by this rollout that have the desired template spec.
	UpdatedReplicas int32
	// Total number of ready pods targeted by this rollout.
	ReadyReplicas int32
	// Total number of available pods (ready for at least minReadySeconds) targeted by this rollout.
	AvailableReplicas int32
	// CurrentStepIndex defines the current step of the rollout is on. If the current step index is null, the
	// controller will execute the rollout.
	CurrentStepIndex int32
	// Conditions a list of conditions a rollout can have.
	Conditions []*RolloutCondition
	// HPAReplicas the number of non-terminated replicas that are receiving active traffic
	HPAReplicas int32
	// Selector that identifies the pods that are receiving active traffic
	Selector string
	// StableRS indicates the replicaset that has successfully rolled out
	StableRS string
	// RestartedAt indicates last time a Rollout was restarted
	RestartedAt time.Time
	// PromoteFull indicates if the rollout should perform a full promotion, skipping analysis and pauses.
	PromoteFull bool
}

type KubernetesAtomicDeployment struct {
	Name         string
	Type         string
	ResolvedName string
	Status       string // TODO enum
	State        *KubernetesState
}

type KubernetesDeploymentState struct {
	Atomic []*KubernetesAtomicDeployment
}

type KubernetesDeployment struct {
	Id              string
	StartedAt       string
	Status          string // TODO enum
	InitiatedBy     string
	InitiatedMethod string
	Env             *Environment
	Application     string
	State			*KubernetesDeploymentState
}