package cli

// wellKnownChannels is a curated list of cloud-native and DevOps conference channels.
// Each entry is "handle\tdisplay name" for shell tab-completion hints.
var wellKnownChannels = []string{
	// CNCF & Kubernetes ecosystem
	"@cncf\tCloud Native Computing Foundation",
	"@KubernetesKubernetes\tKubernetes",
	"@KubernetesCommunity\tKubernetes Community",
	"@linuxfoundation\tLinux Foundation",
	"@LinuxfoundationOrg\tLinux Foundation (org)",
	"@BackstageCommunity\tBackstage Community",
	// Service mesh & networking
	"@Istio\tIstio",
	"@isovalent\tIsovalent (Cilium)",
	"@ciliumproject\tCilium Project",
	"@eBPFCilium\teBPF & Cilium",
	"@envoyproxy\tEnvoy Proxy",
	"@Crossplane\tCrossplane",
	// Observability
	"@grafana\tGrafana Labs",
	"@opentelemetry\tOpenTelemetry",
	"@isitobservable\tIs It Observable",
	// GitOps & deployments
	"@argoproject\tArgo Project",
	"@fluxcd\tFlux CD",
	"@DaprDev\tDapr",
	// Cloud providers
	"@awsdevelopers\tAWS Developers",
	"@AWSEventsChannel\tAWS Events",
	"@amazonwebservices\tAmazon Web Services",
	"@googlecloud\tGoogle Cloud",
	"@Google\tGoogle",
	"@GoogleCloudTech\tGoogle Cloud Tech",
	"@GoogleDevelopers\tGoogle Developers",
	"@GoogleOpenSource\tGoogle Open Source",
	"@GoogleTechTalks\tGoogle Tech Talks",
	"@GoogleCloudEvents\tGoogle Cloud Events",
	"@MicrosoftAzure\tMicrosoft Azure",
	// HashiCorp & infrastructure
	"@hashicorp\tHashiCorp",
	// Security
	"@AquaSecOSS\tAqua Security OSS",
	"@OktaDev\tOkta Developer",
	"@authzed\tAuthzed (SpiceDB)",
	// Containers & platform
	"@DockerInc\tDocker",
	"@ContainerDays\tContainerDays",
	"@ContainersfromtheCouch\tContainers from the Couch",
	"@vmwarecloudnativeapps816\tVMware Cloud Native Apps",
	"@PlatformEngineering\tPlatform Engineering",
	"@kubesimplify\tKube Simplify",
	// Go language
	"@GolangChannel\tGolang Channel",
	"@golang\tGo Language",
	// Tech media & conferences
	"@DevOpsToolkit\tDevOps Toolkit",
	"@thoughtworks\tThoughtWorks",
	"@ByteByteGo\tByte Byte Go",
	"@GOTO\tGOTO Conferences",
	"@infoq\tInfoQ",
	"@DevoxxForever\tDevoxx",
	"@KodeKloud\tKodeKloud",
	"@wiggitywhitney\tWhitney Lee",
}
