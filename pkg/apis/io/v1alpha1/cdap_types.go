package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceType is the name identifying various CDAP master services
type ServiceType string

const (
	AppFabricService ServiceType = "AppFabric"
	LogsProcessorService ServiceType = "LogsProcessor"
	LogsQueryService ServiceType = "LogsQuery"
	MessagingService ServiceType = "Messaging"
	MetadataService ServiceType = "Metadata"
	MetricsProcessorService ServiceType = "MetricsProcessor"
	MetricsQueryService ServiceType = "MetricsQuery"
	UserInterfaceService ServiceType = "UserInterface"
	RouterService ServiceType = "Router"
)

// CDAPSpec defines the desired state of CDAP
type CDAPSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Image    string        `json:"image"`
	Services []CDAPService `json:"services"`
}

// CDAPService defines specification for one CDAP system service
type CDAPService struct {
	Type      ServiceType                  `json:"type"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	Instances *int32                       `json:"instances,omitempty"`
}

// CDAPStatus defines the observed state of CDAP
type CDAPStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CDAP is the Schema for the cdaps API
// +k8s:openapi-gen=true
type CDAP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CDAPSpec   `json:"spec,omitempty"`
	Status CDAPStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CDAPList contains a list of CDAP
type CDAPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CDAP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CDAP{}, &CDAPList{})
}
