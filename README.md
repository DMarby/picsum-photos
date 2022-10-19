Lorem Picsum
===========

Lorem Ipsum... but for photos.  
Lorem Picsum is a service providing easy to use, stylish placeholders.  
It's written in Go, and uses Redis and DigitalOcean Spaces.

## Running locally for development
First, make sure you have Go installed, and this git repo cloned.  
You will also need to [install libvips](https://libvips.github.io/libvips/install.html).

To build the frontend, you need to have NodeJS installed.
Run the following commands to install the dependencies and build it:
```
cd ./web
npm install
npm run-script build
```
If you want to automatically rebuild when you make changes while developing, you can use `npm run-script watch`. 

Then, to start the app, with an in-memory cache, and the test fixtures for images, run:
```
go run . -log-level debug
```
This will start a server accessible on `localhost:8080`, with debug logging enabled.  
For other options/backends, see `go run . -h`.  

Instructions on how to add pictures are available [below](#3-adding-pictures).
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

### 2. Kubernetes
Picsum runs on top of DigitalOcean's hosted Kubernetes offering.

Install [doctl](https://github.com/digitalocean/doctl) and log in with the API token you created earlier by running `doctl auth init`  
Run `doctl kubernetes cluster kubeconfig save picsum-k8s` to set up the kubernetes configuration. Note that you need to have `kubectl` installed.  
Then, run `kubectl config use-context do-ams3-picsum-k8s` to switch to the new configuration.  

#### Secrets
To give the Picsum application access to various things we need to create secrets in Kubernetes.

First, we need to create another spaces access key for the app to access Spaces.
Go to API -> Tokens/Keys in the DigitalOcean control panel, and click "Generate New Key". Copy the Key and the secret.
Then, we'll add it to kubernetes, along with the name of the space, and the region, that we defined earlier in `terraform.tfvars`.
```
kubectl create secret generic picsum-spaces --from-literal=space='SPACE_HERE' --from-literal=access_key='ACCESS_KEY_HERE' --from-literal=secret_key='SECRET_KEY_HERE' --from-literal=endpoint='https://REGION_HERE.digitaloceanspaces.com'
```

Then, we need to create a hmac key that the different services will use to authenticate the requests between eachother:
```
kubectl create secret generic picsum-hmac --from-literal=hmac_key="$(printf '%s' $(pwgen -s 64 1))"
```

#### HTTPS
You'll need to upload an SSL certificate that the cluster will use for https. For picsum.photos, we use a Cloudflare origin certificate.

First, edit `kubernetes/ingress.yaml` and replace the picsum domains with your own domains.  
Then, upload your certificate and private key to the cluster:
```
kubectl create secret tls picsum-cert --key ${KEY_FILE} --cert ${CERT_FILE}
```

You'll also need to configure Picsum so that it knows what domains to use. 
Edit `kubernetes/picsum.yaml` and add the following to the `env` section:
```
- name: PICSUM_ROOT_URL
  value: "https://example.com"
- name: PICSUM_IMAGE_SERVICE_URL
  value: "https://i.example.com"
```

#### DNS
We use Cloudflare to manage our DNS, and as our CDN.  
If you want to have the cluster automatically update your domain to point towards your loadbalancer, you need to configure `external-dns`.  
You may also skip this step if you prefer to manage the DNS manually, simply add an A record for your domain that points towards the loadbalancer IP.

First, create a new API token in Cloudflare, with the following settings:  

Permissions:
- Zone, Zone, Read
- Zone, DNS, Edit

Zone Resources:
- Include, All Zones

Then, run the command below to add the API token to kubernetes:
```
kubectl create secret generic external-dns --from-literal=cf_api_token='API_TOKEN_HERE'
```

Note that you will need to manually set up a CNAME for the domain you specified for the image-service (`i.example.com`) that points towards your main domain (`example.com`).

#### Deployment
Then, go to the `kubernetes` directory and run the following command to create the kubernetes deployment:
```
helmfile apply --environment production
```
Note that this requires `helm`, `helmfile` and `helm-diff` to be installed.

By default, the ingress requires being behind Cloudflare with Authenticated Origin Pulls enabled.
To disable this, set `cloudflareAuthEnabled` to `false` in `kubernetes/environments/production.yaml`.

Now everything should be running, and you should be able to access your instance of Picsum by going to `https://your-domain-pointing-to-the-loadbalancer`.  
Note that the loadbalancer/cluster *only* serves https.


### 3. Adding pictures
To add pictures for the service to use, they need to be added to both spaces, as well as to the `image-manifest.yaml` kubernetes manifest.

#### Spaces
In the DigitalOcean control panel, go to Spaces -> picsum-photos and upload your pictures.  
They should be named `{id}.jpg`, eg `foo.jpg`.

## License
MIT. See [LICENSE](./LICENSE.md)

