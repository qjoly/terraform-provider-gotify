terraform {
  required_providers {
    gotify = {
      source  = "terraform.local/local/gotify"
      version = "0.0.1"
    }
  }
}

provider "gotify" {}

resource "gotify_application" "flux" {
  description = "Je veux une belle description pour mon application"
  name = "super-nom"
}