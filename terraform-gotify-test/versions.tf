terraform {
  required_providers {
    gotify = {
      source  = "terraform.local/local/gotify"
      version = "0.0.1"
    }
  }
}

provider "gotify" {
  token = "Cw6OrEP3AZ5tuGr"
  url = "http://localhost:8080"
}

resource "gotify_application" "app1" {
  description = "Je veux une nouvelle description pour mon application"
  name = "app1"
  priority = "3"
}

resource "gotify_application" "app2" {
  description = "Je veux une nouvelle description pour mon application"
  name = "app2"
}