terraform {
  required_version = ">= 1.3.0"

  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.26.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.17.0"
    }

    fastly = {
      source  = "fastly/fastly"
      version = "3.0.4"
    }
  }
}
