terraform {
  required_version = ">= 1.6.0"

  required_providers {
    scaleway = {
      source  = "scaleway/scaleway"
      version = "~> 2.88"
    }
  }
}

provider "scaleway" {
  project_id = var.project_id
  region     = var.region
}
