package serviceprincipals_test

import (
	"context"
	"fmt"
	"github.com/manicminer/hamilton/msgraph"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azuread/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azuread/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azuread/internal/clients"
	"github.com/hashicorp/terraform-provider-azuread/internal/services/serviceprincipals/parse"
	"github.com/hashicorp/terraform-provider-azuread/internal/utils"
)

type SynchronizationJobProvisionOnDemandResource struct{}

func TestAccSynchronizationJobProvisionOnDemand_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_synchronization_job_provision_on_demand", "test")
	r := SynchronizationJobProvisionOnDemandResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("template_id").Exists(),
				check.That(data.ResourceName).Key("enabled").HasValue("true"),
			),
		},
	})
}

func (r SynchronizationJobProvisionOnDemandResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	client := clients.ServicePrincipals.SynchronizationJobClient
	client.BaseClient.DisableRetries = true

	id, err := parse.SynchronizationJobID(state.ID)
	if err != nil {
		return nil, fmt.Errorf("parsing synchronization job ID: %v", err)
	}

	synchronizationProvisionOnDemand := &msgraph.SynchronizationJobProvisionOnDemand{
		Parameters: &[]msgraph.SynchronizationJobApplicationParameters{
			{
				RuleId: utils.String("//TODO"),
				Subjects: &[]msgraph.SynchronizationJobSubject{
					{
						ObjectId:       utils.String("//TODO get azuread_group.test.id"),
						ObjectTypeName: utils.String("Group"),
					},
				},
			},
		},
	}
	status, err := client.ProvisionOnDemand(ctx, id.JobId, synchronizationProvisionOnDemand, id.ServicePrincipalId)
	if err != nil {
		return nil, fmt.Errorf("retrieving Provision on demand job with object ID %q errored %s", id.JobId, err)
	}
	switch status {
	case http.StatusCreated:
		return nil, fmt.Errorf("Provision on demand job %q did not run provision group %q", id.JobId, id.ServicePrincipalId)
	case http.StatusOK:
		return utils.Bool(true), nil
	default:
		return nil, fmt.Errorf("Provision on demand job %q was not found for service principal %q", id.JobId, id.ServicePrincipalId)
	}
}

func (SynchronizationJobProvisionOnDemandResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}

data "azuread_client_config" "test" {}

data "azuread_application_template" "test" {
  display_name = "Azure Databricks SCIM Provisioning Connector"
}

resource "azuread_application" "test" {
  display_name = "acctestSynchronizationJob-%[1]d"
  owners       = [data.azuread_client_config.test.object_id]
  template_id  = data.azuread_application_template.test.template_id
}

resource "azuread_service_principal" "test" {
  application_id = azuread_application.test.application_id
  owners         = [data.azuread_client_config.test.object_id]
  use_existing   = true
}

resource "azuread_synchronization_job" "test" {
  service_principal_id = azuread_service_principal.test.id
  template_id          = "dataBricks"
}

resource "azuread_group" "test" {
  display_name     = "acctestGroup-%[1]d"
  security_enabled = true
}
`, data.RandomInteger)
}

func (r SynchronizationJobProvisionOnDemandResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_synchronization_job_provision_on_demand" "test" {
  service_principal_id = azuread_service_principal.test.id
  job_id 			   = azuread_synchronization_job.test.id
  parameters {
    rule_id = //TODO 
    subjects {
      object_id        = azuread_group.test.id
      object_type_name = "Group"
    }
  }
}

`, r.template(data))
}
