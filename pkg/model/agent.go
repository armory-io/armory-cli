package model

type Agent struct {
	AgentIdentifier        string `json:"agentIdentifier,omitempty" yaml:"agentIdentifier,omitempty"`
	AgentInstanceUUID      string `json:"agentInstanceUuid,omitempty" yaml:"agentInstanceUuid,omitempty"`
	AgentVersion           string `json:"agentVersion,omitempty" yaml:"agentVersion,omitempty"`
	ClientID               string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ConnectedAtIso8601     string `json:"connectedAtIso8601,omitempty" yaml:"connectedAtIso8601,omitempty"`
	EnvID                  string `json:"envId,omitempty" yaml:"envId,omitempty"`
	HubInstanceUUID        string `json:"hubInstanceUuid,omitempty" yaml:"hubInstanceUuid,omitempty"`
	IPAddress              string `json:"ipAddress,omitempty" yaml:"ipAddress,omitempty"`
	K8sClusterRoleSupport  bool   `json:"k8sClusterRoleSupport,omitempty" yaml:"k8sClusterRoleSupport,omitempty"`
	LastHeartbeatAtIso8601 string `json:"lastHeartbeatAtIso8601,omitempty" yaml:"lastHeartbeatAtIso8601,omitempty"`
	NodeIP                 string `json:"nodeIp,omitempty" yaml:"nodeIp,omitempty"`
	OrgID                  string `json:"orgId,omitempty" yaml:"orgId,omitempty"`
	StreamID               string `json:"streamId,omitempty" yaml:"streamId,omitempty"`
}
