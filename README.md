<<<<<<< HEAD
# Mattermost Genesis

Mattermost Genesis is a tool meant to smooth Mattermost Cloud enterprise adoption. It provides a service to create isolated AWS accounts, provision the networking inrastructure and prepare the ground for Mattermost cloud cluster creation. It offers CIDR pool storage functionality and future releases will include features like VPC peering automation.

## Overview

The following features are supported currently by Genesis:
- Creation a CIDR pool from provided parent subnets.
- Creation of AWS accounts.
- Provisioning of AWS accounts with all necessary infrastructure:
    - VPCs, Subnets, Route tables
    - TGWs, NATs, DHCP options
    - Security Groups
    - DB parameter groups
    - IAM resources
- Provisioning with preselected CIDR or random picked one.
- Deletion of infrastructure and AWS accounts.
- Ability to list and get accounts, subnets, parentsubnets.

The following features will be added to Genesis later:
- Genesis VPC peering feature.
- Automatic registration of new accounts with Mattermost OU groups to enable SSO fast login.
- Ability to clean account from resources without deleting it.
- Integration with Mattermost cloud provisioner.

## Developing

### Environment Setup

#### Required Software

The following is required to properly run the Genesis server.

##### Note: when versions are specified, it is extremely important to follow the requirement. Newer versions will often not work as expected

1. Install [Go](https://golang.org/doc/install)
2. Install [Terraform](https://learn.hashicorp.com/terraform/getting-started/install.html) version v0.13.5
3. Install [golang/mock](https://github.com/golang/mock#installation) version 1.4.x

#### Other Setup

1. Specify the region in your AWS config, e.g. `~/.aws/config`:
```
[profile genesis]
region = us-east-1
```
2. Generate an AWS Access and Secret key pair, then export them in your bash profile:
  ```
  export AWS_ACCESS_KEY_ID=YOURACCESSKEYID
  export AWS_SECRET_ACCESS_KEY=YOURSECRETACCESSKEY
  export AWS_PROFILE=genesis
  ```

3. Create an S3 bucket to store the terraform state.
  ```bash
  aws s3api create-bucket --bucket terraform-genesis-state-bucket-<env> --region us-east-1
  ```

4. Clone this repository into your GOPATH (or anywhere if you have Go Modules enabled)

### Building

Simply run the following:

```bash
go install ./cmd/genesis
alias genesis='$HOME/go/bin/genesis'
```


### Running
Before running the server the first time you must set up the DB with:

```bash
$ genesis schema migrate
```

Run the server with all required flags. See list below:

```
genesis server
--control-tower-account <the account id of the account tha manages control tower>
--control-tower-role <the iam role that will be used in the control tower account>
--managed-ou <the name of the OU used for new accounts>
--sso-first-name <the first name of the SSO user>
--sso-last-name <the last name of the SSO user>
--sso-user-email <the email of the SSO user>
--core-account <the account id of the account that manages the transit gateway>
--resource-share-id <the resource share ID, used to share TGW with new generated accounts>
--bind-ips <the bind servers that will be used for dns resolution>
--cnc-cidrs <the cidr ranges of the cnc subnets that provisioner runs>
--state-bucket <the terraform state bucket>
--tgw-id  <the Transit Gateway ID of the TGW that new VPCs will be attached to>
--tgw-routes <the routes that will be used for TGW traffic>
--teleport-cidr <the teleport CIDR to allow teleport access>
```

All cidr and route ranges should passed in the following format:

i.e.
```
--bind-ips '["10.0.0.0/8", "10.0.0.0/8"]'
```

In a different terminal/window, to add a parent subnet that will be used to provision a subnet pool:
```bash
genesis parent-subnet add --cidr <cidr-range> --split-range <split-range>
i.e.
genesis parent-subnet add --cidr "10.80.0.0/12" --split-range 24
```
You will get a response like this one:
```bash
{
    "ID": "rrmo366shjrofx9utyz8rnwmxr",
    "CIDR": "10.80.0.0/12",
    "SplitRange": 24,
    "CreateAt": 1617794511340,
    "LockAcquiredBy": null,
    "LockAcquiredAt": 0
}
```

By running this command a subnet pool of /24 range was created using the provided parent subnet. You can list the subnet pool CIDRs by running

```bash
genesis subnet list
```
where you can also pass the --table flag to list in a table and the --free-subnets flag to get all non used subnets.

To create a new AWS account you can run:

```bash
genesis account create --service-catalog-product <product-id>
```
You will get a response like this one:
```bash
{
    "ID": "6rk5dxsbrjygbewooninyqzfuy",
    "State": "creation-requested",
    "Provider": "aws",
    "ProviderMetadataAWS": {
        "ServiceCatalogProductID": "prod-xxxx",
        "AWSAccountID": "",
        "AccountProductID": ""
    },
    "AccountMetadata": {
        "Provision": true,
        "Subnet": ""
    },
    "Provisioner": "genesis",
    "CreateAt": 1617789384716,
    "DeleteAt": 0,
    "APISecurityLock": false,
    "LockAcquiredBy": null,
    "LockAcquiredAt": 0
}
```

Check its creation progress on the first window where the API runs or run `genesis account list` to check cluster status.

In the creation step if `--provision` flag is added the account will be provisioned with all necessary infrastructure after its creation. If no subnet is specified with `--subnet` flag a random subnet will be picked from the subnet pool.

If something breaks and account reprovisioning is needed, run
```bash
genesis account provision --account <account-ID>
i.e.
genesis account provision --account 6rk5dxsbrjygbewooninyqzfuy
```


To provision an empty account with a specific subnet cidr you can run:

```bash
genesis ccount provision --account <account-ID> --subnet <subnet-CIDR>
```

### Deleting an account and deployed infrastructure

Before deleting an account and the infrastructure provisioned by Genesis you will **have** to delete the clusters, databases and anything that was created on top of Genesis provisioning.

At the moment account deletion, deletes first all the Terraform deployed infrastructure and then deletes the account. In the future, a feature will be added to empty the account but keep it running hot for future deployments.

To proceed with account deletion:

```bash
genesis account delete --account <account-ID>
i.e.
genesis account delete --account 6rk5dxsbrjygbewooninyqzfuy
```
=======
# Genesis

This repository houses the open-source components of Mattermost Genesis. This tool is providing a service to provision AWS Accounts and manage VPC CIDR pools, VPC infrastructure and VPC peering functionality.
>>>>>>> 9995a40a6838b599341e4e6f16c6ebe992ada984
