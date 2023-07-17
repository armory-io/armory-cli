package pipelineRequest

import "list"
import "struct"

#PipelineRequest: {
  version: string | *"v1"
  kind: string | *"kubernetes"
  application: string
  deploymentConfig?: #DeploymentConfig
  targets: [string]: #DeploymentTarget
  targets: struct.MinFields(1)
  targets: [string]: { strategy: or(_strategyNames) }
  targets: [string]: {
    constraints: #Constraints & { dependsOn?: [... or(_targetNames)]}
  }
  _targetNames: [ for k, v in targets {k}]
  manifests: [... #ManifestPath & { targets?: [... or(_targetNames)] }]
  manifests: list.MinItems(1)
  strategies: #Strategies
  _strategyNames: [ for k, v in strategies {k}]
  analysis?: #PipelineAnalysisConfig
  webhooks?: [... #WebhookConfig]
  trafficManagement?: [... #TrafficManagement & { targets?: [... or(_targetNames)] }]
  files?: #Files
  contextOverrides?: [string]: string
  targetFilters?: [... #IncludeTargetByName & { includeTarget: or(_targetNames) }]
  targetFilters?: list.MinItems(1)
}

#TimeUnit: =~ "(?i)^none|seconds|minutes|hours$"
#RollMode: =~ "(?i)^automatic|manual$"
#LookbackMethod: =~ "(?i)^growing|sliding$"

#DeploymentConfig: {
  timeout?: #Timeout
  keepDeploymentObject?: bool
}

#Timeout: {
  duration: >=1
  unit: #TimeUnit
}

#DeploymentTarget: {
  account: string
  namespace?: string
  strategy?: string
  constraints?: #Constraints
}

#Constraints: {
  dependsOn?: [... string]
  dependsOn?: list.MinItems(1)
  beforeDeployment?: [... #BeforeDeploymentStep]
  afterDeployment?: [... #AfterDeploymentStep]
}

#BeforeDeploymentStep: #PauseStep | #WebhookStep
#AfterDeploymentStep: #PauseStep | #WebhookStep | #AnalysisStep
#CanaryStep: #WeightStep | #PauseStep | #WebhookStep | #AnalysisStep | #ExposeServices
#BlueGreenCondition: #PauseStep | #WebhookStep | #AnalysisStep | #ExposeServices

#ExposeServices: {
  exposeServices: {
    services: [... string]
    services:  list.MinItems(1)
    ttl?: #Timeout
  }
}

#PauseStep: {
  pause: {
    untilApproved: bool
    requiresRoles?: [... string]
    requiresRoles?: list.MinItems(1)
    approvalExpiration?: #Timeout
  } | {
    duration: int
    unit: #TimeUnit
  }

}

#WebhookStep: {
  runWebhook: {
    name: string
    context?: [string]: string
  }
}

#WeightStep: {
  setWeight: {
    weight: int
  }
}

#AnalysisStep: {
  analysis: {
    context?: [string]: string
    rollBackMode: #RollMode
    rollForwardMode: #RollMode
    interval: >=1
    units: #TimeUnit
    numberOfJudgmentRuns: >=1
    queries: [... string]
    queries: list.MinItems(1)
    lookbackMethod: #LookbackMethod
    abortOnFailedJudgment?:bool
    metricProviderName?: string
  }
}

#ManifestPath: {
  #PathOrInline
  targets?: [... string]
}

#PathOrInline: {path: string} | {inline: string}

#Strategies: [string]: #Strategy

#Strategy: #PipelineCanaryStrategy | #PipelineBlueGreenStrategy

#PipelineCanaryStrategy: {
  canary: {
    steps: [... #CanaryStep]
  }
}

#PipelineBlueGreenStrategy: {
  blueGreen: {
    redirectTrafficAfter?: [... #BlueGreenCondition]
    shutDownOldVersionAfter?: [... #BlueGreenCondition]
    activeService?: string
    previewService?: string
  }
}

#PipelineAnalysisConfig: {
  defaultMetricProviderName?: string
  queries: [... #Query]
}

#Query: this={
  name: string
  queryTemplate: string
  upperLimit?: float
  lowerLimit?: float
  #AnyOfLimits: true & list.MinItems([ for label, _ in this if list.Contains(["upperLimit", "lowerLimit"], label) {label}], 1)
  metricProviderName?: string
}

#WebhookConfig: {
  name: string
  method?: string
  uriTemplate: string
  networkMode?: "direct" | "remoteNetworkAgent"
  isRemoteNetworkAgent: bool | *false
  if networkMode != _|_ {
    isRemoteNetworkAgent: networkMode == "remoteNetworkAgent"
  }
  if isRemoteNetworkAgent == true {
    agentIdentifier: string
  }
  headers?: [... #Header]
  bodyTemplate?: #Body
  retryCount?: int
  disableCallback?: bool
}

#Header: {
  key: string
  value: string
}

#Body: {
  inline: string
}

#Files: [string]: [... string]

#TrafficManagement: {
  targets?: [... string]
  smi?: [... #SmiTrafficManagementConfig]
  kubernetes?: [... #KubernetesTrafficManagementConfig]
  istio?: [... #IstioTrafficManagementConfig]
}

#KubernetesTrafficManagementConfig: {
  activeService: string
  previewService?: string
}

#SmiTrafficManagementConfig: {
  rootServiceName: string
  canaryServiceName?: string
  trafficSplitName?:  string
}

#IstioTrafficManagementConfig: {
  virtualService: #IstioVirtualServiceConfig
  destinationRule: #IstioDestinationRuleConfig
}

#IstioVirtualServiceConfig: {
  name: string
  httpRouteName?: string
}

#IstioDestinationRuleConfig: {
  name: string
  activeSubsetName?: string
  canarySubsetName?: string
}

#IncludeTargetByName: {
	includeTarget: string
}
