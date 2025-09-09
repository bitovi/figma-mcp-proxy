# Installation and Usage
## Prereqs
1. Create a Keypair using AWS EC2 Key Pairs
    - (Recommendation): Use the following naming format: figma-proxy-<client-name>-<environment>

## Initialize the Terraform

```bash
terraform init -backend-config="key=figma-mcp-proxy/<client>/terraform/staging.tfstate"

terraform apply \
  -var "aws_region=us-east-2" \
    -var "client_name=<client-name>" \
    -var "target_environment=<environment>" \
    -var "acm_certificate_arn=arn:aws:acm:us-east-2:767397775295:certificate/c90a939f-9e92-4556-98a4-09b0f9df430b" \
    -var "hosted_zone_id=Z088938213M784NAAX7NY" \
    -var "key_name=<key-pair name>" \
    -var "private_key_path=<key-pair name>.pem"
```

## Use the output and RDP onto the Windows machine

```bash
terraform output
terraform output -raw administrator_password
```

1. Copy the administrator_password, if there is a trailing `%`, ignore it.
2. Start a remote desktop connection to the instance
    - Using the value of `public_dns`
3. Open a terminal and run the following:
    - cd C://
    - git clone https://github.com/bitovi/figma-mcp-proxy.git
3. [Install Figma](https://www.figma.com/download/desktop/win)
    - Log in
    - Turn on Dev Mode MCP Server
5. Start the Figma-Proxy using the startup script
    - `$env:API_KEY='<api key>'; $env:EXTERNAL_DNS_NAME='<fqdn from the terraform output>'; & go run main.go`

# FAQ
## How can I recreate the Windows Server if I need to?

```bash
terraform taint aws_instance.win2025
# Then run terraform apply as normal
```

## I cannot connect to the windows server
Add your IP address to the [infra/allowed_cidrs.txt](IP Allowlist)




