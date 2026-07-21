# Oracle Cloud Infrastructure Golang SDK (MODULAR)

This is a fork of https://github.com/oracle/oci-go-sdk

This fork is special: all the packages are modules.

The code of the modules are inside the branch [`modules`](https://github.com/nrdcg/oci-go-sdk/tree/modules).

Note: The minimum Go version is go1.21.0 and direct dependencies are up to date.

## Maintenance

The script to update the fork is `update.sh`.

The detection of the new version is done automatically, but it's to use the env vars `SRC_BASE_MAJOR_VERSION` and `SRC_BASE_VERSION` inside the script.

I update the branch "manually" by calling the script for each release of `github.com/oracle/oci-go-sdk`.

## Usage

Note: it's important to use the consistent versions for modules.

### Example

With the official repository:
```go
import (
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/dns"
```

With the fork:
```go
import (
	"github.com/nrdcg/oci-go-sdk/common/v1065"
	"github.com/nrdcg/oci-go-sdk/common/v1065/auth"
	"github.com/nrdcg/oci-go-sdk/dns/v1065"
)
```

## Modules

The modules exist since v65.95.0 (v1065.95.0).

<!-- module list -->

- `github.com/nrdcg/oci-go-sdk/accessgovernancecp/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/adm/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/aidataplatform/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/aidocument/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/ailanguage/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/aispeech/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/aivision/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/analytics/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/announcementsservice/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/apiaccesscontrol/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/apigateway/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/apiplatform/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/apmconfig/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/apmcontrolplane/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/apmsynthetics/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/apmtraces/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/appmgmtcontrol/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/artifacts/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/audit/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/autoscaling/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/bastion/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/batch/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/bds/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/blockchain/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/budget/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/capacitymanagement/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/certificatesmanagement/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/certificates/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/cims/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/cloudbridge/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/cloudguard/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/cloudmigrations/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/clusterplacementgroups/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/common/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/computecloudatcustomer/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/computeinstanceagent/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/containerengine/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/containerinstances/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/containerregistry/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/core/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/costad/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/dashboardservice/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/databasemanagement/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/databasemigration/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/databasetoolsruntime/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/databasetools/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/database/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/datacatalog/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/datacc/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/dataflow/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/dataintegration/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/datalabelingservicedataplane/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/datalabelingservice/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/datasafe/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/datascience/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/dblm/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/dbmulticloud/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/delegateaccesscontrol/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/demandsignal/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/desktops/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/devops/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/dif/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/disasterrecovery/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/distributeddatabase/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/dns/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/emaildataplane/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/email/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/emwarehouse/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/events/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/filestorage/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/fleetappsmanagement/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/fleetsoftwareupdate/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/functions/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/fusionapps/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/gdp/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/generativeaiagentruntime/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/generativeaiagent/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/generativeaidata/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/generativeaiinference/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/generativeai/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/genericartifactscontent/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/goldengate/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/governancerulescontrolplane/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/healthchecks/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/helpers/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/identitydataplane/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/identitydomains/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/identity/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/integration/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/iot/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/jmsjavadownloads/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/jmsutils/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/jms/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/keymanagement/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/licensemanager/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/limitsincrease/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/limits/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/loadbalancer/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/lockbox/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/loganalytics/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/loggingingestion/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/loggingsearch/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/logging/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/lustrefilestorage/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/managedkafka/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/managementagent/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/managementdashboard/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/marketplaceprivateoffer/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/marketplacepublisher/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/marketplace/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/mediaservices/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/mngdmac/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/modeldeployment/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/monitoring/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/multicloud/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/mysql/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/networkfirewall/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/networkloadbalancer/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/nosql/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/objectstorage/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/oce/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/ocicontrolcenter/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/ocvp/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/oda/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/onesubscription/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/ons/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/opa/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/opensearch/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/operatoraccesscontrol/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/opsi/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/optimizer/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/osmanagementhub/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/ospgateway/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/osubbillingschedule/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/osuborganizationsubscription/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/osubsubscription/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/osubusage/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/psa/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/psql/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/queue/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/recovery/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/redis/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/resourceanalytics/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/resourcemanager/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/resourcescheduler/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/resourcesearch/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/rover/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/sch/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/secrets/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/securityattribute/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/self/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/servicecatalog/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/servicemanagerproxy/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/stackmonitoring/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/streaming/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/tenantmanagercontrolplane/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/threatintelligence/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/usageapi/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/usage/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/vault/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/vbsinst/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/visualbuilder/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/vnmonitoring/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/vulnerabilityscanning/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/waas/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/waa/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/waf/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/wlms/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/workrequests/v1065 v1065.121.1`
- `github.com/nrdcg/oci-go-sdk/zpr/v1065 v1065.121.1`

<!-- end module list -->

## Extra Scripts

The extra scripts:
- `modules.sh`: (not used) It allows converting the original repository to a modular repository
- `tags.sh`: (not used) It allows tagging the original repository to a modular repository

## References

- https://github.com/oracle/oci-go-sdk/issues/255
- https://github.com/oracle/oci-go-sdk/issues/348
