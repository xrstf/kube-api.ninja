// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package swaggerdumper

// This is a list of resource names for Kubernetes 1.5 and 1.6,
// whose swagger specs do not yet contain x-kubernetes metadata.

// maps plural => Kind
var resourceNames = map[string]string{
	"apiservices":                         "APIService",
	"certificatesigningrequests":          "CertificateSigningRequest",
	"clusterrolebindings":                 "ClusterRoleBinding",
	"clusterroles":                        "ClusterRole",
	"componentstatuses":                   "ComponentStatus",
	"configmaps":                          "ConfigMap",
	"controllerrevisions":                 "ControllerRevision",
	"cronjobs":                            "CronJob",
	"daemonsets":                          "DaemonSet",
	"deployments":                         "Deployment",
	"endpoints":                           "Endpoints",
	"events":                              "Event",
	"externaladmissionhookconfigurations": "ExternalAdmissionHookConfiguration",
	"horizontalpodautoscalers":            "HorizontalPodAutoscaler",
	"ingresses":                           "Ingress",
	"initializerconfigurations":           "InitializerConfiguration",
	"jobs":                                "Job",
	"limitranges":                         "LimitRange",
	"namespaces":                          "Namespace",
	"networkpolicies":                     "NetworkPolicy",
	"nodes":                               "Node",
	"persistentvolumeclaims":              "PersistentVolumeClaim",
	"persistentvolumes":                   "PersistentVolume",
	"poddisruptionbudgets":                "PodDisruptionBudget",
	"podpresets":                          "PodPreset",
	"pods":                                "Pod",
	"podsecuritypolicies":                 "PodSecurityPolicy",
	"podtemplates":                        "PodTemplate",
	"replicasets":                         "ReplicaSet",
	"replicationcontrollers":              "ReplicationController",
	"resourcequotas":                      "ResourceQuota",
	"rolebindings":                        "RoleBinding",
	"roles":                               "Role",
	"scheduledjobs":                       "ScheduledJob",
	"secrets":                             "Secret",
	"selfsubjectaccessreviews":            "SelfSubjectAccessReview",
	"serviceaccounts":                     "ServiceAccount",
	"services":                            "Service",
	"statefulsets":                        "StatefulSet",
	"storageclasses":                      "StorageClass",
	"subjectaccessreviews":                "SubjectAccessReview",
	"thirdpartyresources":                 "ThirdPartyResource",
	"tokenreviews":                        "TokenReview",
}
