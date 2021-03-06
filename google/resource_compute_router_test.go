package google

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccComputeRouter_basic(t *testing.T) {
	t.Parallel()

	testId := acctest.RandString(10)
	resourceRegion := "europe-west1"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeRouterDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccComputeRouterBasic(testId, resourceRegion),
			},
			resource.TestStep{
				ResourceName:      "google_compute_router.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccComputeRouter_noRegion(t *testing.T) {
	t.Parallel()

	testId := acctest.RandString(10)
	providerRegion := "us-central1"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeRouterDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccComputeRouterNoRegion(testId, providerRegion),
			},
			resource.TestStep{
				ResourceName:      "google_compute_router.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccComputeRouter_full(t *testing.T) {
	t.Parallel()

	testId := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeRouterDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccComputeRouterFull(testId),
			},
			resource.TestStep{
				ResourceName:      "google_compute_router.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccComputeRouter_update(t *testing.T) {
	t.Parallel()

	testId := acctest.RandString(10)
	region := getTestRegionFromEnv()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeRouterDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccComputeRouterBasic(testId, region),
			},
			resource.TestStep{
				ResourceName:      "google_compute_router.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				Config: testAccComputeRouterFull(testId),
			},
			resource.TestStep{
				ResourceName:      "google_compute_router.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				Config: testAccComputeRouterBasic(testId, region),
			},
			resource.TestStep{
				ResourceName:      "google_compute_router.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckComputeRouterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	routersService := config.clientCompute.Routers

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "google_compute_router" {
			continue
		}

		project, err := getTestProject(rs.Primary, config)
		if err != nil {
			return err
		}

		region, err := getTestRegion(rs.Primary, config)
		if err != nil {
			return err
		}

		name := rs.Primary.Attributes["name"]

		_, err = routersService.Get(project, region, name).Do()

		if err == nil {
			return fmt.Errorf("Error, Router %s in region %s still exists",
				name, region)
		}
	}

	return nil
}

func testAccComputeRouterBasic(testId, resourceRegion string) string {
	return fmt.Sprintf(`
		resource "google_compute_network" "foobar" {
			name = "router-test-%s"
			auto_create_subnetworks = false
		}
		resource "google_compute_subnetwork" "foobar" {
			name = "router-test-subnetwork-%s"
			network = "${google_compute_network.foobar.self_link}"
			ip_cidr_range = "10.0.0.0/16"
			region = "%s"
		}
		resource "google_compute_router" "foobar" {
			name = "router-test-%s"
			region = "${google_compute_subnetwork.foobar.region}"
			network = "${google_compute_network.foobar.name}"
			bgp {
				asn = 64514
			}
		}
	`, testId, testId, resourceRegion, testId)
}

func testAccComputeRouterNoRegion(testId, providerRegion string) string {
	return fmt.Sprintf(`
		resource "google_compute_network" "foobar" {
			name = "router-test-%s"
			auto_create_subnetworks = false
		}
		resource "google_compute_subnetwork" "foobar" {
			name = "router-test-subnetwork-%s"
			network = "${google_compute_network.foobar.self_link}"
			ip_cidr_range = "10.0.0.0/16"
			region = "%s"
		}
		resource "google_compute_router" "foobar" {
			name = "router-test-%s"
			network = "${google_compute_network.foobar.name}"
			bgp {
				asn = 64514
			}
		}
	`, testId, testId, providerRegion, testId)
}

func testAccComputeRouterFull(testId string) string {
	return fmt.Sprintf(`
		resource "google_compute_network" "foobar" {
			name = "router-test-%s"
			auto_create_subnetworks = false
		}

		resource "google_compute_router" "foobar" {
			name = "router-test-%s"
			network = "${google_compute_network.foobar.name}"
			bgp {
				asn = 64514
				advertise_mode = "CUSTOM"
				advertised_groups = ["ALL_SUBNETS"]
				advertised_ip_ranges {
					range = "1.2.3.4"
				}
				advertised_ip_ranges {
					range = "6.7.0.0/16"
				}
			}
		}
	`, testId, testId)
}
