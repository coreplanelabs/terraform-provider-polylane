package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"polylane": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	// The skeleton provider talks to no live API, so there is nothing to
	// require yet. Once the Polylane API client lands, fail here when
	// POLYLANE_API_KEY is not set.
}

// TestAccProviderConfigure exercises the full Terraform plan/apply cycle with
// an explicitly configured provider block and no resources.
func TestAccProviderConfigure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "polylane" {
  endpoint = "https://api.polylane.example/api/v1"
}
`,
			},
		},
	})
}
