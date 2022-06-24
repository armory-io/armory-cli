package model

type OrchestrationConfig struct {
	Version           string                       `yaml:"version,omitempty"`
	Kind              string                       `yaml:"kind,omitempty"`
	DeploymentConfig  *DeploymentConfig            `yaml:"deploymentConfig,omitempty"`
	Application       string                       `yaml:"application,omitempty"`
	Targets           *map[string]DeploymentTarget `yaml:"targets,omitempty"`
	Manifests         *[]ManifestPath              `yaml:"manifests,omitempty"`
	Strategies        *map[string]Strategy         `yaml:"strategies,omitempty"`
	Analysis          *AnalysisConfig              `yaml:"analysis,omitempty"`
	Webhooks          *[]WebhookConfig             `yaml:"webhooks,omitempty"`
	TrafficManagement *[]TrafficManagement         `yaml:"trafficManagement,omitempty"`
}

type DeploymentConfig struct {
	Timeout *TimeoutConfig `yaml:"timeout,omitempty"`
}

type TimeoutConfig struct {
	Duration int32  `yaml:"duration,omitempty"`
	Unit     string `yaml:"unit,omitempty"`
}

type Strategy struct {
	Canary    *CanaryStrategy    `yaml:"canary,omitempty"`
	BlueGreen *BlueGreenStrategy `yaml:"blueGreen,omitempty"`
}

type CanaryStrategy struct {
	Steps *[]CanaryStep `yaml:"steps,omitempty"`
}

type CanaryStep struct {
	SetWeight  *WeightStep   `yaml:"setWeight,omitempty"`
	Pause      *PauseStep    `yaml:"pause,omitempty"`
	Analysis   *AnalysisStep `yaml:"analysis,omitempty"`
	RunWebhook *WebhookStep  `yaml:"runWebhook,omitempty"`
}

type BlueGreenStrategy struct {
	RedirectTrafficAfter    []*BlueGreenCondition `yaml:"redirectTrafficAfter,omitempty"`
	ShutDownOldVersionAfter []*BlueGreenCondition `yaml:"shutDownOldVersionAfter,omitempty"`
	ActiveService           string                `yaml:"activeService,omitempty"`
	PreviewService          string                `yaml:"previewService,omitempty"`
}

type BlueGreenCondition struct {
	Pause      *PauseStep    `yaml:"pause,omitempty"`
	Analysis   *AnalysisStep `yaml:"analysis,omitempty"`
	RunWebhook *WebhookStep  `yaml:"runWebhook,omitempty"`
}

type WeightStep struct {
	Weight int32 `yaml:"weight,omitempty"`
}

type PauseStep struct {
	// The duration of the pause. If duration is non-zero, untilApproved should be set to false.
	Duration int32  `yaml:"duration,omitempty"`
	Unit     string `yaml:"unit,omitempty"`
	// If set to true, the progressive canary will wait until a manual judgment to continue. This field should not be set to true unless duration and unit are unset.
	UntilApproved bool `yaml:"untilApproved,omitempty"`
}

type AnalysisStep struct {
	Context               map[string]string `yaml:"context,omitempty"`
	RollBackMode          string            `yaml:"rollBackMode,omitempty"`
	RollForwardMode       string            `yaml:"rollForwardMode,omitempty"`
	Interval              int32             `yaml:"interval,omitempty"`
	Units                 string            `yaml:"units,omitempty"`
	NumberOfJudgmentRuns  int32             `yaml:"numberOfJudgmentRuns,omitempty"`
	Queries               []string          `yaml:"queries,omitempty"`
	LookbackMethod        string            `yaml:"lookbackMethod,omitempty"`
	AbortOnFailedJudgment bool              `yaml:"abortOnFailedJudgment,omitempty"`
	MetricProviderName    string            `yaml:"metricProviderName,omitempty"`
}

type DeploymentTarget struct {
	// The name of the Kubernetes account to be used for this deployment.
	Account string `yaml:"account,omitempty"`
	// The Kubernetes namespace where the provided manifests will be deployed.
	Namespace string `yaml:"namespace,omitempty"`
	// This is the key to a strategy under the strategies map
	Strategy    string       `yaml:"strategy,omitempty"`
	Constraints *Constraints `yaml:"constraints,omitempty"`
}

type AnalysisConfig struct {
	DefaultMetricProviderName string  `yaml:"defaultMetricProviderName,omitempty"`
	Queries                   []Query `yaml:"queries,omitempty"`
}

type Query struct {
	Name               string  `yaml:"name,omitempty"`
	QueryTemplate      string  `yaml:"queryTemplate,omitempty"`
	AggregationMethod  string  `yaml:"aggregationMethod,omitempty"`
	UpperLimit         *int32  `yaml:"upperLimit,omitempty"`
	LowerLimit         *int32  `yaml:"lowerLimit,omitempty"`
	MetricProviderName *string `yaml:"metricProviderName,omitempty"`
}

type ManifestPath struct {
	Path    string   `yaml:"path,omitempty"`
	Targets []string `yaml:"targets,omitempty"`
	Inline  string   `yaml:"inline,omitempty"`
}

type Constraints struct {
	DependsOn        *[]string           `yaml:"dependsOn,omitempty"`
	BeforeDeployment *[]BeforeDeployment `yaml:"beforeDeployment,omitempty"`
	AfterDeployment  *[]AfterDeployment  `yaml:"afterDeployment,omitempty"`
}

type BeforeDeployment struct {
	Pause      *PauseStep   `yaml:"pause,omitempty"`
	RunWebhook *WebhookStep `yaml:"runWebhook,omitempty"`
}

type AfterDeployment struct {
	Pause      *PauseStep    `yaml:"pause,omitempty"`
	RunWebhook *WebhookStep  `yaml:"runWebhook,omitempty"`
	Analysis   *AnalysisStep `yaml:"analysis,omitempty"`
}

type WebhookConfig struct {
	Name            string    `yaml:"name,omitempty"`
	Method          string    `yaml:"method,omitempty"`
	UriTemplate     string    `yaml:"uriTemplate,omitempty"`
	NetworkMode     string    `yaml:"networkMode,omitempty"`
	AgentIdentifier string    `yaml:"agentIdentifier,omitempty"`
	Headers         *[]Header `yaml:"headers,omitempty"`
	BodyTemplate    *Body     `yaml:"bodyTemplate,omitempty"`
	RetryCount      int32     `yaml:"retryCount,omitempty"`
}

type Header struct {
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

type Body struct {
	Inline *string `yaml:"inline,omitempty"`
	Path   *string `yaml:"path,omitempty"`
}

type WebhookStep struct {
	Name    string            `yaml:"name,omitempty"`
	Context map[string]string `yaml:"context,omitempty"`
}

type TrafficManagement struct {
	Targets    []string            `yaml:"targets,omitempty"`
	SMI        []SMI               `yaml:"smi,omitempty"`
	Kubernetes []KubernetesTraffic `yaml:"kubernetes,omitempty"`
}

type SMI struct {
	RootServiceName   string `yaml:"rootServiceName,omitempty"`
	CanaryServiceName string `yaml:"canaryServiceName,omitempty"`
	TrafficSplitName  string `yaml:"trafficSplitName,omitempty"`
}

type KubernetesTraffic struct {
	ActiveService  string `yaml:"activeService,omitempty"`
	PreviewService string `yaml:"previewService,omitempty"`
}
