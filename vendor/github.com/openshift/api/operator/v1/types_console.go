package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/api/config/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Console struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec ConsoleSpec `json:"spec,omitempty"`
	// +optional
	Status ConsoleStatus `json:"status,omitempty"`
}

type ConsoleSpec struct {
	OperatorSpec `json:",inline"`
	// customization is used to optionally provide a small set of
	// customization options to the web console.
	// +optional
	Customization ConsoleCustomization `json:"customization"`
}

type ConsoleStatus struct {
	OperatorStatus `json:",inline"`
}

type ConsoleCustomization struct {
	// brand is the default branding of the web console which can be overridden by
	// providing the brand field.  There is a limited set of specific brand options.
	// This field controls elements of the console such as the logo.
	// Invalid value will prevent a console rollout.
	Brand Brand `json:"brand,omitempty"`
	// documentationBaseURL links to external documentation are shown in various sections
	// of the web console.  Providing documentationBaseURL will override the default
	// documentation URL.
	// Invalid value will prevent a console rollout.
	DocumentationBaseURL string `json:"documentationBaseURL,omitempty"`
	// customProductName is the name that will be displayed in page titles, logo alt text, and the about dialog
	// instead of the normal OpenShift product name.
	// +optional
	CustomProductName string `json:"customProductName,omitempty"`
	// customLogoFile replaces the default OpenShift logo in the masthead and about dialog. It is a reference to a
	// ConfigMap in the openshift-config namespace. This can be created with a command like
	// 'oc create configmap custom-logo --from-file=/path/to/file -n openshift-config'.
	// Image size must be less than 1 MB due to constraints on the ConfigMap size.
	// The ConfigMap key should include a file extension so that the console serves the file
	// with the correct MIME type.
	// Recommended logo specifications:
	// Dimensions: Max height of 68px and max width of 200px
	// SVG format preferred
	// +optional
	CustomLogoFile v1.ConfigMapFileReference `json:"customLogoFile,omitempty"`
}

// Brand is a specific supported brand within the console
type Brand string

const (
	// Branding for OpenShift
	BrandOpenShift Brand = "openshift"
	// Branding for The Origin Community Distribution of Kubernetes
	BrandOKD Brand = "okd"
	// Branding for OpenShift Online
	BrandOnline Brand = "online"
	// Branding for OpenShift Container Platform
	BrandOCP Brand = "ocp"
	// Branding for OpenShift Dedicated
	BrandDedicated Brand = "dedicated"
	// Branding for Azure Red Hat OpenShift
	BrandAzure Brand = "azure"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ConsoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Console `json:"items"`
}
