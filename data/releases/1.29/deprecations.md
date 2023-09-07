The **v1.29** release will stop serving the following deprecated API versions:

#### Flow control resources {#flowcontrol-resources-v129}

The **flowcontrol.apiserver.k8s.io/v1beta2** API version of FlowSchema and PriorityLevelConfiguration will no longer be served in v1.29.

* Migrate manifests and API clients to use the **flowcontrol.apiserver.k8s.io/v1beta3** API version, available since v1.26.
* All existing persisted objects are accessible via the new API
* Notable changes in **flowcontrol.apiserver.k8s.io/v1beta3**:
  * The PriorityLevelConfiguration `spec.limited.assuredConcurrencyShares` field is renamed to `spec.limited.nominalConcurrencyShares`
