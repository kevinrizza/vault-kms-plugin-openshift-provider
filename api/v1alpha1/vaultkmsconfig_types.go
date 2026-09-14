/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// VaultKMSConfigSpec defines the desired state of VaultKMSConfig.
type VaultKMSConfigSpec struct {
	// vaultAddress specifies the address of the HashiCorp Vault instance.
	// The value must be a valid HTTPS URL containing only scheme, host, and optional port.
	// Paths, user info, query parameters, and fragments are not allowed.
	//
	// Format: https://hostname[:port]
	// Example: https://vault.example.com:8200
	//
	// The value must be between 1 and 512 characters.
	//
	// +kubebuilder:validation:XValidation:rule="isURL(self)",message="must be a valid URL"
	// +kubebuilder:validation:XValidation:rule="isURL(self) && url(self).getScheme() == 'https'",message="must use the 'https' scheme"
	// +kubebuilder:validation:XValidation:rule="isURL(self) && (url(self).getEscapedPath() == '' || url(self).getEscapedPath() == '/')",message="must not contain a path"
	// +kubebuilder:validation:XValidation:rule="isURL(self) && url(self).getQuery() == {}",message="must not have a query"
	// +kubebuilder:validation:XValidation:rule="self.find('#(.+)$') == ''",message="must not have a fragment"
	// +kubebuilder:validation:XValidation:rule="self.find('@') == ''",message="must not have user info"
	// +kubebuilder:validation:MaxLength=512
	// +kubebuilder:validation:MinLength=1
	// +required
	VaultAddress string `json:"vaultAddress,omitempty"`

	// vaultNamespace specifies the Vault namespace where the Transit secrets engine is mounted.
	// This is only applicable for Vault Enterprise installations.
	// When this field is not set, no namespace is used.
	//
	// The value must be between 1 and 4096 characters.
	// The namespace cannot end with a forward slash, cannot contain spaces, and cannot be one of the reserved strings: root, sys, audit, auth, cubbyhole, or identity.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	// +kubebuilder:validation:XValidation:rule="!self.endsWith('/')",message="vaultNamespace cannot end with a forward slash"
	// +kubebuilder:validation:XValidation:rule="!self.contains(' ')",message="vaultNamespace cannot contain spaces"
	// +kubebuilder:validation:XValidation:rule="!(self in ['root', 'sys', 'audit', 'auth', 'cubbyhole', 'identity'])",message="vaultNamespace cannot be a reserved string (root, sys, audit, auth, cubbyhole, identity)"
	// +optional
	VaultNamespace string `json:"vaultNamespace,omitempty"`

	// vaultAuthNamespace specifies the Vault namespace to use for authentication.
	// This is only applicable for Vault Enterprise installations where authentication
	// and Transit operations may be in different namespaces.
	// When this field is not set, the value of vaultNamespace is used for both
	// authentication and Transit key operations.
	//
	// The value must be between 1 and 4096 characters.
	// The namespace cannot end with a forward slash, cannot contain spaces, and cannot be one of the reserved strings: root, sys, audit, auth, cubbyhole, or identity.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	// +kubebuilder:validation:XValidation:rule="!self.endsWith('/')",message="vaultAuthNamespace cannot end with a forward slash"
	// +kubebuilder:validation:XValidation:rule="!self.contains(' ')",message="vaultAuthNamespace cannot contain spaces"
	// +kubebuilder:validation:XValidation:rule="!(self in ['root', 'sys', 'audit', 'auth', 'cubbyhole', 'identity'])",message="vaultAuthNamespace cannot be a reserved string (root, sys, audit, auth, cubbyhole, identity)"
	// +optional
	VaultAuthNamespace string `json:"vaultAuthNamespace,omitempty"`

	// tls contains the TLS configuration for connecting to the Vault server.
	// When this field is not set, system default TLS settings are used.
	// +optional
	TLS VaultTLSConfig `json:"tls,omitzero"`

	// authentication defines the authentication method used to authenticate with Vault.
	//
	// +required
	Authentication VaultAuthentication `json:"authentication,omitzero"`

	// vaultKeyPath specifies the full path to the encryption key in Vault's Transit secrets engine,
	// combining the Transit engine mount path and the key name separated by "/keys/".
	// Format: <mount>/keys/<key-name> (e.g., transit/keys/my-key, myteam/transit/keys/production-key).
	//
	// The total path length must be between 8 and 1542 characters.
	// The path cannot start or end with a forward slash, cannot contain consecutive forward slashes,
	// must only contain RFC 3986 unreserved characters (alphanumeric, hyphen, period, underscore, tilde)
	// and forward slashes as path separators, and must not contain "." or ".." path segments.
	// The key name must start and end with an alphanumeric character or underscore, and may contain
	// alphanumeric characters, underscores, hyphens, and periods in the middle.
	//
	// +kubebuilder:validation:MinLength=8
	// +kubebuilder:validation:MaxLength=1542
	// +kubebuilder:validation:XValidation:rule="!self.startsWith('/')",message="vaultKeyPath cannot start with a forward slash"
	// +kubebuilder:validation:XValidation:rule="!self.endsWith('/')",message="vaultKeyPath cannot end with a forward slash"
	// +kubebuilder:validation:XValidation:rule="!self.contains('//')",message="vaultKeyPath cannot contain consecutive forward slashes"
	// +kubebuilder:validation:XValidation:rule="self.matches('^[a-zA-Z0-9._~/-]+$')",message="vaultKeyPath must only contain RFC 3986 unreserved characters (alphanumeric, hyphen, period, underscore, tilde) and forward slashes"
	// +kubebuilder:validation:XValidation:rule="self.split('/').filter(s, s == '.' || s == '..').size() == 0",message="vaultKeyPath must not contain '.' or '..' path segments"
	// +kubebuilder:validation:XValidation:rule=`self.matches('^[a-zA-Z0-9._~-]+(/[a-zA-Z0-9._~-]+)*/keys/[a-zA-Z0-9_]([a-zA-Z0-9_.-]*[a-zA-Z0-9_])?$')`,message="vaultKeyPath must follow the format <mount>/keys/<key-name> where the key name starts and ends with an alphanumeric character or underscore and may contain alphanumeric characters, underscores, hyphens, and periods"
	// +required
	VaultKeyPath string `json:"vaultKeyPath,omitempty"`
}

// VaultTLSConfig contains TLS configuration for connecting to Vault.
// +kubebuilder:validation:MinProperties=1
type VaultTLSConfig struct {
	// caBundle references a ConfigMap in the openshift-config namespace containing
	// the CA certificate bundle used to verify the TLS connection to the Vault server.
	// The referenced ConfigMap must contain the CA bundle in the key "ca-bundle.crt".
	// When this field is not set, the system's trusted CA certificates are used.
	//
	// The namespace for the ConfigMap is openshift-config.
	//
	// Example ConfigMap:
	//   apiVersion: v1
	//   kind: ConfigMap
	//   metadata:
	//     name: vault-ca-bundle
	//     namespace: openshift-config
	//   data:
	//     ca-bundle.crt: |
	//       -----BEGIN CERTIFICATE-----
	//       ...
	//       -----END CERTIFICATE-----
	//
	// +optional
	CABundle VaultConfigMapReference `json:"caBundle,omitzero"`

	// serverName specifies the Server Name Indication (SNI) to use when connecting to Vault via TLS.
	// This is useful when the Vault server's hostname doesn't match its TLS certificate.
	// When this field is not set, the hostname from vaultAddress is used for SNI.
	//
	// The value must be a valid DNS hostname: it must contain no more than 253 characters,
	// contain only lowercase alphanumeric characters, '-' or '.', and start and end with an alphanumeric character.
	//
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:XValidation:rule="self.matches('^[a-z0-9]([a-z0-9\\\\-]*[a-z0-9])?(\\\\.[a-z0-9]([a-z0-9\\\\-]*[a-z0-9])?)*$')",message="serverName must be a valid DNS hostname: contain no more than 253 characters, contain only lowercase alphanumeric characters, '-' or '.', and start and end with an alphanumeric character"
	// +optional
	ServerName string `json:"serverName,omitempty"`
}

// VaultAuthentication defines the authentication method used to authenticate with Vault.
// +kubebuilder:validation:XValidation:rule="self.type == 'AppRole' ? has(self.appRole) : !has(self.appRole)",message="appRole config is required when authentication type is AppRole, and forbidden otherwise"
// +union
type VaultAuthentication struct {
	// type defines the authentication method used to authenticate with Vault.
	// Allowed values are AppRole.
	// When set to AppRole, the plugin uses AppRole credentials to authenticate with Vault.
	//
	// +unionDiscriminator
	// +required
	Type VaultAuthenticationType `json:"type,omitempty"`

	// appRole defines the configuration for AppRole authentication.
	// This field must be set when type is AppRole, and must be unset otherwise.
	//
	// +unionMember
	// +optional
	AppRole VaultAppRoleAuthentication `json:"appRole,omitzero"`
}

// VaultAuthenticationType defines the authentication method type for Vault.
// +kubebuilder:validation:Enum=AppRole
type VaultAuthenticationType string

const (
	// VaultAuthenticationTypeAppRole represents AppRole authentication method.
	VaultAuthenticationTypeAppRole VaultAuthenticationType = "AppRole"
)

// VaultAppRoleAuthentication defines the configuration for AppRole authentication with Vault.
type VaultAppRoleAuthentication struct {
	// secret references a secret in the openshift-config namespace containing
	// the AppRole credentials used to authenticate with Vault.
	// The referenced Secret must contain two keys: "role-id" for the AppRole Role ID and "secret-id" for the AppRole Secret ID.
	//
	// +required
	Secret VaultSecretReference `json:"secret,omitzero"`
}

// VaultSecretReference references a secret in the openshift-config namespace.
type VaultSecretReference struct {
	// name is the metadata.name of the referenced secret in the openshift-config namespace.
	// The name must be a valid DNS subdomain name: it must contain no more than 253 characters,
	// contain only lowercase alphanumeric characters, '-' or '.', and start and end with an alphanumeric character.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:XValidation:rule="self.matches('^[a-z0-9]([a-z0-9\\\\-]*[a-z0-9])?(\\\\.[a-z0-9]([a-z0-9\\\\-]*[a-z0-9])?)*$')",message="name must be a valid DNS subdomain name: contain no more than 253 characters, contain only lowercase alphanumeric characters, '-' or '.', and start and end with an alphanumeric character"
	// +required
	Name string `json:"name,omitempty"`
}

// VaultConfigMapReference references a ConfigMap in the openshift-config namespace.
type VaultConfigMapReference struct {
	// name is the metadata.name of the referenced ConfigMap in the openshift-config namespace.
	// The name must be a valid DNS subdomain name: it must contain no more than 253 characters,
	// contain only lowercase alphanumeric characters, '-' or '.', and start and end with an alphanumeric character.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:XValidation:rule="self.matches('^[a-z0-9]([a-z0-9\\\\-]*[a-z0-9])?(\\\\.[a-z0-9]([a-z0-9\\\\-]*[a-z0-9])?)*$')",message="name must be a valid DNS subdomain name: contain no more than 253 characters, contain only lowercase alphanumeric characters, '-' or '.', and start and end with an alphanumeric character"
	// +required
	Name string `json:"name,omitempty"`
}

// VaultKMSConfigStatus defines the observed state of VaultKMSConfig.
type VaultKMSConfigStatus struct {
	// kmsPluginImage is the resolved container image for the HashiCorp Vault KMS plugin,
	// set by the controller. This may be a tag-based or digest-based image reference.
	//
	// +optional
	KMSPluginImage string `json:"kmsPluginImage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type VaultKMSConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultKMSConfigSpec   `json:"spec,omitempty"`
	Status VaultKMSConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VaultKMSConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultKMSConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(GroupVersion, &VaultKMSConfig{}, &VaultKMSConfigList{})
		return nil
	})
}
