terraform {
  required_providers {
    gotify = {
      source  = "terraform.local/local/gotify"
      version = "0.0.1"
    }
  }
}

provider "gotify" {}

resource "gotify_feed" "flux" {
  configurable_attribute = "some-value"

}