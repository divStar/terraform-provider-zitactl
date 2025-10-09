# Add Testcontainers for Automated Testing

## Docker Compose Configuration

File: `docker-compose.test.yaml`
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U zitadel"]
      interval: 5s
      timeout: 5s
      retries: 5

  zitadel:
    image: ghcr.io/zitadel/zitadel:v2.62.1
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "8443:8443"
    volumes:
      - ./testdata/machine-keys:/machine-keys
    environment:
      ZITADEL_EXTERNALDOMAIN: localhost
      ZITADEL_EXTERNALPORT: 8443
      ZITADEL_EXTERNALSECURE: "true"
      ZITADEL_TLS_ENABLED: "true"
      ZITADEL_FIRSTINSTANCE_ORG_NAME: TestOrg
      ZITADEL_FIRSTINSTANCE_ORG_HUMAN_USERNAME: zitadel-admin
      ZITADEL_FIRSTINSTANCE_ORG_HUMAN_FIRSTNAME: Test
      ZITADEL_FIRSTINSTANCE_ORG_HUMAN_LASTNAME: Admin
      ZITADEL_FIRSTINSTANCE_ORG_HUMAN_EMAIL_ADDRESS: admin@test.local
      ZITADEL_FIRSTINSTANCE_ORG_HUMAN_EMAIL_VERIFIED: "true"
      ZITADEL_FIRSTINSTANCE_ORG_HUMAN_PASSWORD: TestPassword123!
      ZITADEL_FIRSTINSTANCE_ORG_MACHINE_MACHINE_USERNAME: zitadel-admin-sa
      ZITADEL_FIRSTINSTANCE_ORG_MACHINE_MACHINE_NAME: AdminServiceAccount
      ZITADEL_FIRSTINSTANCE_ORG_MACHINE_MACHINEKEY_TYPE: 1
      ZITADEL_FIRSTINSTANCE_MACHINEKEYPATH: /machine-keys/zitadel-admin-sa.json
      ZITADEL_DATABASE_POSTGRES_HOST: postgres
      ZITADEL_DATABASE_POSTGRES_PORT: 5432
      ZITADEL_DATABASE_POSTGRES_DATABASE: zitadel
      ZITADEL_DATABASE_POSTGRES_USER_USERNAME: zitadel
      ZITADEL_DATABASE_POSTGRES_USER_PASSWORD: zitadel
      ZITADEL_DATABASE_POSTGRES_USER_SSL_MODE: disable
      ZITADEL_DATABASE_POSTGRES_ADMIN_USERNAME: postgres
      ZITADEL_DATABASE_POSTGRES_ADMIN_PASSWORD: postgres
      ZITADEL_DATABASE_POSTGRES_ADMIN_SSL_MODE: disable
```
## Go Test Setup with Testcontainers

File: `internal/provider/org/data_source_container_test.go`

```go
package org_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/divStar/terraform-provider-zitactl/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestContainersWithCompose(t *testing.T) (cleanup func()) {
	t.Helper()
	
	ctx := context.Background()
	
	// Ensure machine-keys directory exists
	os.MkdirAll("./testdata/machine-keys", 0755)
	
	composeFilePaths := []string{"docker-compose.test.yaml"}
	
	compose, err := compose.NewDockerCompose(composeFilePaths...)
	if err != nil {
		t.Fatalf("Failed to create compose: %v", err)
	}

	// Start containers
	err = compose.
		WaitForService("zitadel", 
			wait.ForHTTP("/debug/healthz").
				WithPort("8080/tcp").
				WithStartupTimeout(5*time.Minute)).
		Up(ctx, compose.Wait(true))
	
	if err != nil {
		t.Fatalf("Failed to start compose: %v", err)
	}

	// Wait additional time for Zitadel to fully initialize and generate machine key
	t.Log("Waiting for Zitadel to generate machine key...")
	time.Sleep(2 * time.Minute)

	// Read the generated machine key
	keyPath := filepath.Join(".", "testdata", "machine-keys", "zitadel-admin-sa.json")
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		compose.Down(ctx, compose.RemoveOrphans(true))
		t.Fatalf("Failed to read machine key: %v", err)
	}

	// Set environment variables for tests
	os.Setenv("ZITACTL_DOMAIN", "localhost:8443")
	os.Setenv("ZITACTL_SKIP_TLS_VERIFICATION", "true")
	os.Setenv("ZITACTL_SERVICE_ACCOUNT_KEY", string(keyBytes))
	os.Setenv("ZITACTL_TEST_ORG_NAME", "TestOrg")

	cleanup = func() {
		os.Unsetenv("ZITACTL_DOMAIN")
		os.Unsetenv("ZITACTL_SKIP_TLS_VERIFICATION")
		os.Unsetenv("ZITACTL_SERVICE_ACCOUNT_KEY")
		os.Unsetenv("ZITACTL_TEST_ORG_NAME")
		compose.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
	}

	return cleanup
}

func TestAccOrgsDataSource_WithContainers(t *testing.T) {
	if os.Getenv("USE_TESTCONTAINERS") != "1" {
		t.Skip("Testcontainer tests - set USE_TESTCONTAINERS=1 to run")
	}

	cleanup := setupTestContainersWithCompose(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgsDataSourceConfig("TestOrg"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.zitactl_orgs.test", "name", "TestOrg"),
					resource.TestCheckResourceAttrSet("data.zitactl_orgs.test", "ids.#"),
				),
			},
		},
	})
}
```

## Running the Tests

```bash
# Regular acceptance tests (requires real Zitadel instance)
export TF_ACC=1
export ZITACTL_DOMAIN="zitadel.my.world"
export ZITACTL_SERVICE_ACCOUNT_KEY='{"type":"serviceaccount",...}'
go test -v ./internal/provider/org/

# Testcontainer tests (fully automated)
export USE_TESTCONTAINERS=1
export TF_ACC=1
go test -v ./internal/provider/org/ -run TestAccOrgsDataSource_WithContainers
```

## Key points

> [!NOTE]
> The `docker-compose.yml` as well as the scripts and modules will need some tweaking as they are **untested**.

1. Zitadel runs on port `8443` (mapped from container's `8443`)
2. Uses `skip_tls_verification: true` to bypass certificate validation
3. Machine key is written to `./testdata/machine-keys/zitadel-admin-sa.json`
4. Startup takes ~2 minutes for full initialization
5. Testcontainers manages the entire lifecycle (start/stop)