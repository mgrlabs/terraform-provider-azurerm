package azurerm

import (
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/automation/mgmt/2015-10-31/automation"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmAutomationDscNodeConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmAutomationDscNodeConfigurationCreateUpdate,
		Read:   resourceArmAutomationDscNodeConfigurationRead,
		Update: resourceArmAutomationDscNodeConfigurationCreateUpdate,
		Delete: resourceArmAutomationDscNodeConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"automation_account_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"resource_group_name": resourceGroupNameSchema(),

			"content_embedded": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"configuration_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceArmAutomationDscNodeConfigurationCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationDscNodeConfigurationClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureRM Automation Dsc Node Configuration creation.")

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	accName := d.Get("automation_account_name").(string)
	content := d.Get("content_embedded").(string)

	// configuration name is always the first part of the dsc node configuration
	// e.g. webserver.prod or webserver.local will be associated to the dsc configuration webserver

	configurationName := strings.Split(name, ".")[0]

	parameters := automation.DscNodeConfigurationCreateOrUpdateParameters{
		Source: &automation.ContentSource{
			Type:  automation.EmbeddedContent,
			Value: utils.String(content),
		},
		Configuration: &automation.DscConfigurationAssociationProperty{
			Name: utils.String(configurationName),
		},
		Name: utils.String(name),
	}

	_, err := client.CreateOrUpdate(ctx, resourceGroup, accName, name, parameters)
	if err != nil {
		return err
	}

	read, err := client.Get(ctx, resourceGroup, accName, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Automation Dsc Node Configuration %q (resource group %q) ID", name, resourceGroup)
	}

	d.SetId(*read.ID)

	return resourceArmAutomationDscNodeConfigurationRead(d, meta)
}

func resourceArmAutomationDscNodeConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationDscNodeConfigurationClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	accName := id.Path["automationAccounts"]
	name := id.Path["nodeConfigurations"]

	resp, err := client.Get(ctx, resourceGroup, accName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on AzureRM Automation Dsc Node Configuration %q: %+v", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("automation_account_name", accName)
	d.Set("configuration_name", resp.Configuration.Name)

	// cannot read back content_embedded as not part of body nor exposed through method

	return nil
}

func resourceArmAutomationDscNodeConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationDscNodeConfigurationClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	accName := id.Path["automationAccounts"]
	name := id.Path["nodeConfigurations"]

	resp, err := client.Delete(ctx, resourceGroup, accName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return fmt.Errorf("Error issuing AzureRM delete request for Automation Dsc Node Configuration %q: %+v", name, err)
	}

	return nil
}
