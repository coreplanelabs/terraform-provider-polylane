package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/coreplanelabs/terraform-provider-polylane/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"polylane": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("POLYLANE_API_KEY") == "" {
		t.Fatal("POLYLANE_API_KEY must be set for acceptance tests")
	}
}

func testAccWorkspacePreCheck(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	if os.Getenv("POLYLANE_WORKSPACE_ID") == "" {
		t.Skip("POLYLANE_WORKSPACE_ID is not set; skipping live workspace acceptance test")
	}
}

func testAccTeamPreCheck(t *testing.T) {
	t.Helper()
	testAccWorkspacePreCheck(t)

	if os.Getenv("POLYLANE_TEAM_ACCEPTANCE") != "1" {
		t.Skip("POLYLANE_TEAM_ACCEPTANCE=1 is required for the mutating team lifecycle test")
	}
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
  endpoint = "https://api.polylane.example/v1"
}
`,
			},
		},
	})
}

func TestAccWorkspaceImport(t *testing.T) {
	workspaceID := os.Getenv("POLYLANE_WORKSPACE_ID")
	config := fmt.Sprintf(`
provider "polylane" {}

resource "polylane_workspace" "test" {}

import {
  to = polylane_workspace.test
  id = %q
}
`, workspaceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccWorkspacePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("polylane_workspace.test", "id", workspaceID),
					resource.TestCheckResourceAttrSet("polylane_workspace.test", "name"),
					resource.TestCheckResourceAttrSet("polylane_workspace.test", "slug"),
					resource.TestCheckResourceAttrSet("polylane_workspace.test", "owner_id"),
				),
			},
		},
	})
}

func TestAccWorkspaceMemberImport(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env %q set", resource.EnvTfAcc)
	}
	testAccWorkspacePreCheck(t)

	workspaceID := os.Getenv("POLYLANE_WORKSPACE_ID")
	endpoint := os.Getenv("POLYLANE_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	apiClient, err := client.New(os.Getenv("POLYLANE_API_KEY"), endpoint, "test", nil)
	if err != nil {
		t.Fatalf("create API client: %v", err)
	}
	workspace, err := apiClient.GetWorkspace(context.Background(), workspaceID)
	if err != nil {
		t.Fatalf("read workspace to find owner: %v", err)
	}

	config := fmt.Sprintf(`
provider "polylane" {}

resource "polylane_workspace_member" "test" {
  workspace_id = %q
  user_id      = %q
}

import {
  to = polylane_workspace_member.test
  id = %q
}
`, workspaceID, workspace.OwnerID, workspaceID+"/"+workspace.OwnerID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccWorkspacePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("polylane_workspace_member.test", "workspace_id", workspaceID),
					resource.TestCheckResourceAttr("polylane_workspace_member.test", "user_id", workspace.OwnerID),
					resource.TestCheckResourceAttr("polylane_workspace_member.test", "position", "owner"),
					resource.TestCheckResourceAttrSet("polylane_workspace_member.test", "email"),
				),
			},
		},
	})
}

func TestAccTeamLifecycle(t *testing.T) {
	workspaceID := os.Getenv("POLYLANE_WORKSPACE_ID")
	initialName := "Terraform Acceptance " + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	updatedName := initialName + " Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccTeamPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamConfig(workspaceID, initialName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("polylane_team.test", "workspace_id", workspaceID),
					resource.TestCheckResourceAttr("polylane_team.test", "name", initialName),
					resource.TestCheckResourceAttrSet("polylane_team.test", "id"),
					resource.TestCheckResourceAttrSet("polylane_team.test", "slug"),
				),
			},
			{
				Config: testAccTeamConfig(workspaceID, updatedName),
				Check:  resource.TestCheckResourceAttr("polylane_team.test", "name", updatedName),
			},
			{
				ResourceName:      "polylane_team.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					team := state.RootModule().Resources["polylane_team.test"]
					return workspaceID + "/" + team.Primary.Attributes["id"], nil
				},
			},
		},
	})
}

func testAccTeamConfig(workspaceID, name string) string {
	return fmt.Sprintf(`
provider "polylane" {}

resource "polylane_team" "test" {
  workspace_id = %q
  name         = %q
  description  = "Created by the Terraform provider acceptance suite."
}
`, workspaceID, name)
}
