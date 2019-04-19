resource "digitalocean_database_cluster" "picsum-db" {
  name       = "picsum-db"
  engine     = "pg"
  version    = "11"
  size       = "db-s-1vcpu-1gb"
  region     = "ams3"
  node_count = 1
}
