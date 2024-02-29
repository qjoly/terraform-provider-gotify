
terraform {
  required_providers {
    gotify = {
      source  = "terraform.local/local/gotify"
      version = "0.0.1"
    }
  }
}

provider "gotify" {
  token = "CAZMEZi72TLmRCE"
  url = "http://localhost:8080"
}
