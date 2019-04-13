resource "digitalocean_database_cluster" "picsum_cluster" {
  name       = "picsum-cluster"
  engine     = "pg"
  version    = "11"
  size       = "db-s-1vcpu-1gb"
  region     = "ams3"
  node_count = 1
}
