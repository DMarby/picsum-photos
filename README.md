Lorem Picsum
===========

Lorem Ipsum... but for photos.  
Lorem Picsum is a service providing easy to use, stylish placeholders.  
It's written in Go, and uses Redis, PostgreSQL and DigitalOcean Spaces.

## Running locally for development
First, make sure you have Go installed, and this git repo cloned.
To start a local instance using an in-memory cache, and the test fixtures for images, run:
```
go run . -log-level debug
```
This will start a server accessible on `localhost:8080`, with debug logging enabled.  
For other options/backends, see `go run . -h`.  

Information on how to set up the postgres backend can be found [here](#database-migrations).  
Instructions on how to add pictures to the postgres/spaces backends are available [below](#3-adding-pictures).

### Creating new database migrations
In order to create new database migrations if you need to modify the database structure, run:
```
migrate create -ext sql -dir migrations my_new_migration
```
Then add your SQL to `migrations/<timestamp>_my_new_migration.up.sql` and `migrations/<timestamp>_my_new_migration.down.sql`

## Deployment on DigitalOcean
<p>This project is kindly hosted by:</p>
<p>
  <a href="https://www.digitalocean.com/?utm_medium=opensource&utm_source=picsum">
    <img src="https://opensource.nyc3.cdn.digitaloceanspaces.com/attribution/assets/SVG/DO_Logo_horizontal_blue.svg" width="201px">
  </a>
</p>

To deploy your own instance of Picsum on DigitalOcean, start by cloning this repo using git. Then follow the steps below.

### 1. Terraform
This project uses terraform to set up the infrastructure.

To get started, you'll need to create a few things in the DigitalOcean control panel:
- A private DigitalOcean Space for Terraform remote state
  - Go to Create -> Spaces, choose "Restrict File Listing", and select a name.
- A Spaces access key for Terraform to access the space
  - Go to API -> Tokens/Keys, and click "Generate New Key". Copy the Key and the secret.
- An API key for Terraform to access DigitalOcean
  - Go to API -> Tokens/Keys, and click "Generate New Token". Choose both read and write access.

Then, copy the following file, and replace the default values with your credentials/settings:
```
terraform/terraform.tf.example -> terraform/terraform.tf
terraform/terraform.tfvars.example -> terraform/terraform.tfvars
```
Note that the `endpoint` in `terraform.tf` needs to match the region you created the DigitalOcean Space in.

Then, go to the `terraform` directory, and run `terraform init`.  
You can now set up the infrastructure by running `terraform apply`.

### 2. Configuring the database
To start using the database that Terraform created, you need to do some additional setup.

In the DigitalOcean control panel, go to:
- Databases -> picsum-db -> Settings:
  - Add `picsum-k8s-worker` to "Allow inbound sources"
  - Add "Your computers current IP" to "Allow inbound sources"
    - You should remove this once you are one adding images/running migrations against the database.
- Databases -> picsum-db -> Users & Databases:
  - Create a new user named `picsum`
  - Create a new database named `picsum`
- Databases -> picsum-db -> Connection Pools:
  - Create a new connection pool named `picsum`
    - Database: `picsum`
    - User: `picsum`
    - Pool mode: `transaction`
    - Pool size: 22

Get the connection string from the connection details link for the connection pool.

#### Database migrations
Next up, you need to run the migrations to set up the database for the application.
We use [migrate](https://github.com/golang-migrate/migrate) to handle database migrations.
Install it, and then run
```
migrate -path migrations -database 'connection_string_here' up
```
to set up your database.

### 3. Adding pictures
To add pictures for the service to use, they need to be added to both spaces, as well as the database.

#### Spaces
In the DigitalOcean control panel, go to Spaces -> picsum-photos and upload your pictures.  
They should be named `{id}.jpg`, eg `foo.jpg`.

#### Database
Connect to the Postgres database using a postgres client, and add an entry for each image into the `image` table:  
```
insert into image (id, author, url, width, height) VALUES ('foo', 'John Doe', 'https://picsum.photos', 300, 400);
```

### 4. Kubernetes
Picsum runs on top of DigitalOcean's hosted Kubernetes offering.

Install [doctl](https://github.com/digitalocean/doctl) and log in with the API token you created earlier by running `doctl auth init`  
Run `doctl kubernetes cluster kubeconfig save picsum-k8s` to set up the kubernetes configuration. Note that you need to have `kubectl` installed.  
Then, run `kubectl config use-context do-ams3-picsum-k8s` to switch to the new configuration.  

#### Secrets
To give the Picsum application access to various things we need to create secrets in Kubernetes.
First, we need to store the connection string we got earlier for the postgres database connection pool. 
```
kubectl create secret generic picsum-db --from-literal=connection_string='CONNECTION_STRING_HERE'
```

Then, we need to create another spaces access key for the app to access Spaces.
Go to API -> Tokens/Keys in the DigitalOcean control panel, and click "Generate New Key". Copy the Key and the secret.
Then, we'll add it to kubernetes, along with the name of the space, and the region, that we defined earlier in `terraform.tfvars`.
```
kubectl create secret generic picsum-spaces --from-literal=space='SPACE_HERE' --from-literal=region='REGION_HERE' --from-literal=access_key='ACCESS_KEY_HERE' --from-literal=secret_key='SECRET_KEY_HERE'
```

Then, we need another API key for `external-dns`, so that it can update the A record of the domain to point towards the load balancer.
Go to API -> Tokens/Keys in the DigitalOcean control panel, and click "Generate New Token". Choose both read and write access.
Run the command below to add it to kubernetes:
```
kubectl create secret generic external-dns --from-literal=do_token='API_TOKEN_HERE'
```

#### Deployment
Then, go to the `kubernetes` directory and run the following command to create the kubernetes deployment:
```
kubectl apply -f .
```

Finally, we need to annotate the load balancer with the domain to update the A record for, the same one that we defined in `terraform.tfvars`:
```
kubectl annotate service picsum-lb "external-dns.alpha.kubernetes.io/hostname=example.com"
```

Now everything should be running, and you should be able to access your instance of Picsum by going to the domain you defined in `terraform.tfvars` in your browser.

## License
MIT. See [LICENSE](./LICENSE.md)
