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

resource "gotify_application" "app1" {
  description = "Je veux une nouvelle description pour mon application"
  name = "app1"
  priority = "3"
}

resource "gotify_application" "app2" {
  description = "Je veux une nouvelle description pour mon application"
  name = "app2"
}

data "gotify_application" "test" {
  id = "4"
}

# Show output of the data source
output "test" {
  value = data.gotify_application.test
}
