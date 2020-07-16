package v1beta1

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"strings"
)

const (
	ClusterStatusExpanding   = "expanding"
	ClusterStatusOperational = "operational"
	ClusterStatusFailed      = "failed"

	StorageRollingUpdateFailed = "failed to perform rolling update on vmStorage"
	StorageCreationFailed      = "failed to create vmStorage statefulset"

	SelectRollingUpdateFailed = "failed to perform rolling update on vmSelect"
	SelectCreationFailed      = "failed to create vmSelect statefulset"
	InsertCreationFailed      = "failed to create vmInsert deployment"
)

// VmClusterSpec defines the desired state of VMCluster
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Insert Version",type="string",JSONPath=".spec.vminsert.version",description="The version of VMInsert"
// +kubebuilder:printcolumn:name="Storage Version",type="string",JSONPath=".spec.vmstorage.version",description="The version of VMStorage"
// +kubebuilder:printcolumn:name="Select Version",type="string",JSONPath=".spec.vmselect.version",description="The version of VMSelect"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type VmClusterSpec struct {
	// RetentionPeriod in months
	// +kubebuilder:validation:Pattern:="[1-9]+"
	RetentionPeriod string `json:"retentionPeriod"`
	// ReplicationFactor defines how many copies of data make among
	// distinct storage nodes
	// +optional
	ReplicationFactor *int32 `json:"replicationFactor,omitempty"`

	// ImagePullSecrets An optional list of references to secrets in the same namespace
	// to use for pulling images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// +optional
	VMSelect *VMSelect `json:"vmselect,omitempty"`
	// +optional
	VMInsert *VmInsert `json:"vminsert,omitempty"`
	// +optional
	VMStorage *VMStorage `json:"vmstorage,omitempty"`
}

// VMCluster represents a Victoria-Metrics database cluster
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="VMCluster App"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployment,apps"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Statefulset,apps"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Service,v1"
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=vmclusters,scope=Namespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VMCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VmClusterSpec   `json:"spec"`
	Status            VMClusterStatus `json:"status,omitempty"`
}

func (c *VMCluster) AsOwner() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         c.APIVersion,
			Kind:               c.Kind,
			Name:               c.Name,
			UID:                c.UID,
			Controller:         pointer.BoolPtr(true),
			BlockOwnerDeletion: pointer.BoolPtr(true),
		},
	}
}

// Validate validate cluster specification
func (c VMCluster) Validate() error {
	if err := c.Spec.VMStorage.Validate(); err != nil {
		return err
	}
	return nil
}

// VMClusterStatus defines the observed state of VMCluster
type VMClusterStatus struct {
	UpdateFailCount int             `json:"updateFailCount"`
	LastSync        string          `json:"lastSync,omitempty"`
	ClusterStatus   string          `json:"clusterStatus"`
	Reason          string          `json:"reason,omitempty"`
	VMStorage       VMStorageStatus `json:"vmStorage"`
}

type VMStorageStatus struct {
	Status string `json:"status"`
}

//func (c VMCluster) FullStorageName() string {
//	return fmt.Sprintf("%s-%s", c.Name, c.Spec.VMStorage.GetName())
//}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VMClusterList contains a list of VMCluster
type VMClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VMCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VMCluster{}, &VMClusterList{})
}

type VMSelect struct {
	Name string `json:"name,omitempty"`
	// PodMetadata configures Labels and Annotations which are propagated to the VMSelect pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`
	// Image - docker image settings for VMSelect
	// +optional
	Image Image `json:"image,omitempty"`
	// Secrets is a list of Secrets in the same namespace as the VMSelect
	// object, which shall be mounted into the VMSelect Pods.
	// The Secrets are mounted into /etc/vm/secrets/<secret-name>.
	// +optional
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a list of ConfigMaps in the same namespace as the VMSelect
	// object, which shall be mounted into the VMSelect Pods.
	// The ConfigMaps are mounted into /etc/vm/configmaps/<configmap-name>.
	// +optional
	ConfigMaps []string `json:"configMaps,omitempty"`
	// LogFormat for VMSelect to be configured with.
	//default or json
	// +optional
	// +kubebuilder:validation:Enum=default;json
	LogFormat string `json:"logFormat,omitempty"`
	// LogLevel for VMSelect to be configured with.
	// +optional
	// +kubebuilder:validation:Enum=INFO;WARN;ERROR;FATAL;PANIC
	LogLevel string `json:"logLevel,omitempty"`
	// ReplicaCount is the expected size of the VMSelect cluster. The controller will
	// eventually make the size of the running cluster equal to the expected
	// size.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Pod Count"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podCount"
	ReplicaCount *int32 `json:"replicaCount"`
	// Volumes allows configuration of additional volumes on the output Deployment definition.
	// Volumes specified will be appended to other volumes that are generated as a result of
	// StorageSpec objects.
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output Deployment definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the VMSelect container,
	// that are generated as a result of StorageSpec objects.
	// +optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// Resources container resource request and limits, https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Resources"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity If specified, the pod's scheduling constraints.
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// Tolerations If specified, the pod's tolerations.
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	// +optional
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// VMSelect Pods.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Containers property allows to inject additions sidecars. It can be useful for proxies, backup, etc.
	// +optional
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
	// fetch secrets for injection into the VMSelect configuration from external sources. Any
	// errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// Using initContainers for any use case other then secret fetching is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that this behaviour may break
	// at any time without notice.
	// +optional
	InitContainers []v1.Container `json:"initContainers,omitempty"`

	// Priority class assigned to the Pods
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// HostNetwork controls whether the pod may use the node network namespace
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty"`
	// DNSPolicy sets DNS policy for the pod
	// +optional
	DNSPolicy v1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// CacheMountPath allows to add cache persistent for VMSelect
	// +optional
	CacheMountPath string `json:"cacheMountPath,omitempty"`

	// Storage - add persistent volume for cacheMounthPath
	// its useful for persistent cache
	// +optional
	Storage *StorageSpec `json:"persistentVolume,omitempty"`
	// ExtraEnvs that will be added to VMSelect pod
	// +optional
	ExtraEnvs []v1.EnvVar `json:"extraEnvs,omitempty"`
	// +optional
	ExtraArgs map[string]string `json:"extraArgs,omitempty"`

	//Port listen port
	// +optional
	Port string `json:"port,omitempty"`

	// SchedulerName - defines kubernetes scheduler name
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`
}

func (s VMSelect) GetName(clusterName string) string {
	if s.Name == "" {
		return PrefixedName(clusterName, "vmselect")
	}
	return PrefixedName(s.Name, "vmselect")
}
func (s VMSelect) BuildPodNameWithService(baseName string, podIndex int32, namespace string, portName string) string {
	return fmt.Sprintf("%s-%d.%s.%s.svc:%s,", baseName, podIndex, baseName, namespace, portName)
}

func PrefixedName(name, prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, name)
}

type VmSelectPV struct {
	// +optional
	AccessModes []v1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// +optional
	ExistingClaim string `json:"existingClaim,omitempty"`
	// +optional
	Size string `json:"size,omitempty"`
	// +optional
	SubPath string `json:"subPath,omitempty"`
}

type VmInsert struct {
	// +optional
	Name string `json:"name,omitempty"`

	// PodMetadata configures Labels and Annotations which are propagated to the VMSelect pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`

	// Image - docker image settings for VMInsert
	// +optional
	Image Image `json:"image,omitempty"`
	// Secrets is a list of Secrets in the same namespace as the VMSelect
	// object, which shall be mounted into the VMSelect Pods.
	// The Secrets are mounted into /etc/vmalert/secrets/<secret-name>.
	// +optional
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a list of ConfigMaps in the same namespace as the VMSelect
	// object, which shall be mounted into the VMSelect Pods.
	// The ConfigMaps are mounted into /etc/vmalert/configmaps/<configmap-name>.
	// +optional
	ConfigMaps []string `json:"configMaps,omitempty"`
	// LogFormat for VMSelect to be configured with.
	//default or json
	// +optional
	// +kubebuilder:validation:Enum=default;json
	LogFormat string `json:"logFormat,omitempty"`
	// LogLevel for VMSelect to be configured with.
	// +optional
	// +kubebuilder:validation:Enum=INFO;WARN;ERROR;FATAL;PANIC
	LogLevel string `json:"logLevel,omitempty"`
	// ReplicaCount is the expected size of the VMSelect cluster. The controller will
	// eventually make the size of the running cluster equal to the expected
	// size.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Pod Count"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podCount"
	ReplicaCount *int32 `json:"replicaCount"`
	// Volumes allows configuration of additional volumes on the output Deployment definition.
	// Volumes specified will be appended to other volumes that are generated as a result of
	// StorageSpec objects.
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output Deployment definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the VMSelect container,
	// that are generated as a result of StorageSpec objects.
	// +optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// Resources container resource request and limits, https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Resources"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity If specified, the pod's scheduling constraints.
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// Tolerations If specified, the pod's tolerations.
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	// +optional
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// VMSelect Pods.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Containers property allows to inject additions sidecars. It can be useful for proxies, backup, etc.
	// +optional
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
	// fetch secrets for injection into the VMSelect configuration from external sources. Any
	// errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// Using initContainers for any use case other then secret fetching is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that this behaviour may break
	// at any time without notice.
	// +optional
	InitContainers []v1.Container `json:"initContainers,omitempty"`

	// Priority class assigned to the Pods
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// HostNetwork controls whether the pod may use the node network namespace
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty"`
	// DNSPolicy sets DNS policy for the pod
	// +optional
	DNSPolicy v1.DNSPolicy `json:"dnsPolicy,omitempty"`
	// +optional
	ExtraArgs map[string]string `json:"extraArgs,omitempty"`

	//Port listen port
	// +optional
	Port string `json:"port,omitempty"`

	// SchedulerName - defines kubernetes scheduler name
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`
	// ExtraEnvs that will be added to VMSelect pod
	// +optional
	ExtraEnvs []v1.EnvVar `json:"extraEnvs,omitempty"`
}

func (i VmInsert) GetName(clusterName string) string {
	if i.Name == "" {
		return PrefixedName(clusterName, "vminsert")
	}
	return PrefixedName(i.Name, "vminsert")
}

type VMStorage struct {
	// +optional
	Name string `json:"name,omitempty"`

	// PodMetadata configures Labels and Annotations which are propagated to the VMSelect pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`

	// Image - docker image settings for VMStorage
	// +optional
	Image Image `json:"image,omitempty"`

	// Secrets is a list of Secrets in the same namespace as the VMSelect
	// object, which shall be mounted into the VMSelect Pods.
	// The Secrets are mounted into /etc/vmalert/secrets/<secret-name>.
	// +optional
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a list of ConfigMaps in the same namespace as the VMSelect
	// object, which shall be mounted into the VMSelect Pods.
	// The ConfigMaps are mounted into /etc/vmalert/configmaps/<configmap-name>.
	// +optional
	ConfigMaps []string `json:"configMaps,omitempty"`
	// LogFormat for VMSelect to be configured with.
	//default or json
	// +optional
	// +kubebuilder:validation:Enum=default;json
	LogFormat string `json:"logFormat,omitempty"`
	// LogLevel for VMSelect to be configured with.
	// +optional
	// +kubebuilder:validation:Enum=INFO;WARN;ERROR;FATAL;PANIC
	LogLevel string `json:"logLevel,omitempty"`
	// ReplicaCount is the expected size of the VMSelect cluster. The controller will
	// eventually make the size of the running cluster equal to the expected
	// size.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Pod Count"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podCount"
	ReplicaCount *int32 `json:"replicaCount"`
	// Volumes allows configuration of additional volumes on the output Deployment definition.
	// Volumes specified will be appended to other volumes that are generated as a result of
	// StorageSpec objects.
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output Deployment definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the VMSelect container,
	// that are generated as a result of StorageSpec objects.
	// +optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// Resources container resource request and limits, https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Resources"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity If specified, the pod's scheduling constraints.
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// Tolerations If specified, the pod's tolerations.
	// +optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	// +optional
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// VMSelect Pods.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Containers property allows to inject additions sidecars. It can be useful for proxies, backup, etc.
	// +optional
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
	// fetch secrets for injection into the VMSelect configuration from external sources. Any
	// errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// Using initContainers for any use case other then secret fetching is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that this behaviour may break
	// at any time without notice.
	// +optional
	InitContainers []v1.Container `json:"initContainers,omitempty"`

	// Priority class assigned to the Pods
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// HostNetwork controls whether the pod may use the node network namespace
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty"`
	// DNSPolicy sets DNS policy for the pod
	// +optional
	DNSPolicy v1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// StorageDataPath - path to storage data
	// +optional
	StorageDataPath string `json:"storageDataPath,omitempty"`
	// Storage - add persistent volume for StorageDataPath
	// its useful for persistent cache
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`

	// +optional
	TerminationGracePeriodSeconds int64 `json:"terminationGracePeriodSeconds,omitempty"`
	// SchedulerName - defines kubernetes scheduler name
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`

	//Port for health check connetions
	Port string `json:"port,omitempty"`

	// VMInsertPort for VMInsert connections
	// +optional
	VMInsertPort string `json:"vmInsertPort,omitempty"`

	// VMSelectPort for VMSelect connections
	// +optional
	VMSelectPort string `json:"vmSelectPort,omitempty"`

	// +optional
	ExtraArgs map[string]string `json:"extraArgs,omitempty"`
	// ExtraEnvs that will be added to VMSelect pod
	// +optional
	ExtraEnvs []v1.EnvVar `json:"extraEnvs,omitempty"`
}

// Validate validates vmstorage specification
func (s VMStorage) Validate() error {
	// TODO Fix it
	if err := s.Storage.Validate(); err != nil {
		return fmt.Errorf("invalid persitance volime spec for %s:%w", s.Name, err)
	}
	return nil
}

func (s VMStorage) BuildPodNameWithService(baseName string, podIndex int32, namespace string, portName string) string {
	return fmt.Sprintf("%s-%d.%s.%s.svc:%s,", baseName, podIndex, baseName, namespace, portName)
}

func (s VMStorage) GetName(clusterName string) string {
	if s.Name == "" {
		return PrefixedName(clusterName, "vmstorage")
	}
	return PrefixedName(s.Name, "vmstorage")
}

func (s VMStorage) GetStorageName() string {

	return "vmstorage-db"
}

// Validate validates vmstorage specification
func (s VMSelect) GetCacheMountName() string {
	return PrefixedName("cachedir", "vmselect")
}

// Image defines docker image settings
type Image struct {
	// Repository contains name of docker image + it's repository if needed
	Repository string `json:"repository,omitempty"`
	// Tag contains desired docker image version
	Tag string `json:"tag,omitempty"`
	// PullPolicy describes how to pull docker image
	PullPolicy v1.PullPolicy `json:"pullPolicy,omitempty"`
}

func (cr VMCluster) VMSelectSelectorLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "vmselect",
		"app.kubernetes.io/instance":  cr.Name,
		"app.kubernetes.io/component": "monitoring",
		"managed-by":                  "vm-operator",
	}
}

func (cr VMCluster) VMSelectPodLabels() map[string]string {
	labels := cr.VMSelectSelectorLabels()
	if cr.Spec.VMSelect == nil || cr.Spec.VMSelect.PodMetadata == nil {
		return labels
	}
	for label, value := range cr.Spec.VMSelect.PodMetadata.Labels {
		labels[label] = value
	}
	return labels
}

func (cr VMCluster) VMInsertSelectorLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "vminsert",
		"app.kubernetes.io/instance":  cr.Name,
		"app.kubernetes.io/component": "monitoring",
		"managed-by":                  "vm-operator",
	}
}

func (cr VMCluster) VMInsertPodLabels() map[string]string {
	labels := cr.VMInsertSelectorLabels()
	if cr.Spec.VMInsert == nil || cr.Spec.VMInsert.PodMetadata == nil {
		return labels
	}
	for label, value := range cr.Spec.VMInsert.PodMetadata.Labels {
		labels[label] = value
	}
	return labels
}
func (cr VMCluster) VMStorageSelectorLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "vmstorage",
		"app.kubernetes.io/instance":  cr.Name,
		"app.kubernetes.io/component": "monitoring",
		"managed-by":                  "vm-operator",
	}
}

func (cr VMCluster) VMStoragePodLabels() map[string]string {
	labels := cr.VMStorageSelectorLabels()
	if cr.Spec.VMStorage == nil || cr.Spec.VMStorage.PodMetadata == nil {
		return labels
	}
	//TODO fix it, we cannot allow to overwrite selector labels
	for label, value := range cr.Spec.VMStorage.PodMetadata.Labels {
		labels[label] = value
	}
	return labels
}

func (cr VMCluster) FinalLabels(baseLabels map[string]string) map[string]string {
	labels := map[string]string{}
	if cr.ObjectMeta.Labels != nil {
		for label, value := range cr.ObjectMeta.Labels {
			labels[label] = value
		}
	}
	for label, value := range baseLabels {
		labels[label] = value
	}
	return labels
}

func (cr VMCluster) VMSelectPodAnnotations() map[string]string {
	if cr.Spec.VMSelect == nil || cr.Spec.VMSelect.PodMetadata == nil {
		return nil
	}
	return cr.Spec.VMSelect.PodMetadata.Annotations
}

func (cr VMCluster) VMInsertPodAnnotations() map[string]string {
	if cr.Spec.VMInsert == nil || cr.Spec.VMInsert.PodMetadata == nil {
		return nil
	}
	return cr.Spec.VMSelect.PodMetadata.Annotations
}

func (cr VMCluster) VMStoragePodAnnotations() map[string]string {
	if cr.Spec.VMStorage == nil || cr.Spec.VMStorage.PodMetadata == nil {
		return nil
	}
	return cr.Spec.VMSelect.PodMetadata.Annotations
}

func (cr VMCluster) Annotations() map[string]string {
	annotations := make(map[string]string, len(cr.ObjectMeta.Annotations))
	for annotation, value := range cr.ObjectMeta.Annotations {
		if !strings.HasPrefix(annotation, "kubectl.kubernetes.io/") {
			annotations[annotation] = value
		}
	}
	return annotations
}
