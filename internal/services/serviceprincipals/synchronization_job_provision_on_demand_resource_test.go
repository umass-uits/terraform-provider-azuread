package serviceprincipals_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azuread/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azuread/internal/clients"
	"github.com/hashicorp/terraform-provider-azuread/internal/utils"
)

type SynchronizationJobProvisionOnDemandResource struct{}

func TestAccSynchronizationJobProvisionOnDemand_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_synchronization_job_provision_on_demand", "test")
	r := SynchronizationJobProvisionOnDemandResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			//The provisioned app isn't actually integrated so this will never work
			Config:      r.basic(data),
			ExpectError: regexp.MustCompile("CredentialsMissing: Please configure provisioning by providing your admin credentials then retry the provision on-demand."),
		},
	})
}

func (r SynchronizationJobProvisionOnDemandResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	return utils.Bool(true), nil
}

func (SynchronizationJobProvisionOnDemandResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azuread" {}
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-databricks-%d"
  location = "%s"
}

resource "azurerm_databricks_workspace" "test" {
  name                = "acctestDBW-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku                 = "%s"
}

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
  synchronization_job_id = trimprefix(azuread_synchronization_job.test.id, "${azuread_service_principal.test.id}/job/")
  parameter {
	rule_id = "03f7d90d-bf71-41b1-bda6-aaf0ddbee5d8" //no api to check this so assuming the rule id is the same globally :finger_crossed: 
    subject {
      object_id        = azuread_group.test.id
      object_type_name = "Group"
    }
  }
}

`, r.template(data))
}
