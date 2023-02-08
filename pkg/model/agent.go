package model

type Agent struct {
	AgentIdentifier        string `json:"agentIdentifier,omitempty" yaml:"agentIdentifier,omitempty"`
	AgentInstanceUuid      string `json:"agentInstanceUuid,omitempty" yaml:"agentInstanceUuid,omitempty"`
	AgentVersion           string `json:"agentVersion,omitempty" yaml:"agentVersion,omitempty"`
	ClientId               string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ConnectedAtIso8601     string `json:"connectedAtIso8601,omitempty" yaml:"connectedAtIso8601,omitempty"`
	EnvId                  string `json:"envId,omitempty" yaml:"envId,omitempty"`
	HubInstanceUuid        string `json:"hubInstanceUuid,omitempty" yaml:"hubInstanceUuid,omitempty"`
	IpAddress              string `json:"ipAddress,omitempty" yaml:"ipAddress,omitempty"`
	K8sClusterRoleSupport  bool   `json:"k8sClusterRoleSupport,omitempty" yaml:"k8sClusterRoleSupport,omitempty"`
	LastHeartbeatAtIso8601 string `json:"lastHeartbeatAtIso8601,omitempty" yaml:"lastHeartbeatAtIso8601,omitempty"`
	NodeIp                 string `json:"nodeIp,omitempty" yaml:"nodeIp,omitempty"`
	OrgId                  string `json:"orgId,omitempty" yaml:"orgId,omitempty"`
	StreamId               string `json:"streamId,omitempty" yaml:"streamId,omitempty"`
}
