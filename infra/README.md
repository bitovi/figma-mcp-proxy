# Installation and Usage
## Prereqs
1. Set your environment variables
```bash
export CLIENT_NAME=<client name>
export ENVIRONMENT=<staging/production>
export AWS_REGION=us-east-2
``` 

2. Create a Keypair using AWS EC2 Key Pairs
This will create the entry within the Bitovi AI Org key-pairs as well as download the private key to your local machine. (Save this to 1Password for others)
```bash
aws ec2 create-key-pair \
    --key-name "figma-proxy-$CLIENT_NAME-$ENVIRONMENT" \
    --query 'KeyMaterial' \
    --region $AWS_REGION \
    --output text > figma-proxy-$CLIENT_NAME-$ENVIRONMENT.pem
```

## Initialize the Terraform
You shouldn't need to change anything below so long as you've set things up appropriately. You **will. need** to delete the `.terraform` && `.terraform.lock.hcl` folder and file if you're creating a new entry.

```bash
terraform init -backend-config="key=figma-mcp-proxy/$CLIENT_NAME/terraform/$ENVIRONMENT.tfstate"

terraform apply \
  -var "aws_region=us-east-2" \
    -var "client_name=$CLIENT_NAME" \
    -var "target_environment=$ENVIRONMENT" \
    -var "acm_certificate_arn=arn:aws:acm:us-east-2:767397775295:certificate/c90a939f-9e92-4556-98a4-09b0f9df430b" \
    -var "hosted_zone_id=Z088938213M784NAAX7NY" \
    -var "key_name=figma-proxy-$CLIENT_NAME-$ENVIRONMENT" \
    -var "private_key_path=figma-proxy-$CLIENT_NAME-$ENVIRONMENT.pem" \
    -var "proxy_api_token=<api token>"
```

## Use the output and RDP onto the Windows machine

```bash
terraform output
terraform output -raw administrator_password
```

1. Copy the administrator_password, if there is a trailing `%`, ignore it.
2. Start a remote desktop connection to the instance
    - Using the value of `public_dns`
3. [Install Figma](https://www.figma.com/download/desktop/win)
    - Log in
    - Turn on Dev Mode MCP Server
    - **Open any Figma design file** (required for MCP server to work)
4. Setup the Windows Server + Figma Watchdog service (keep-alive daemon)
    `C:\figma-mcp-proxy\setup-figma-alwayson.ps1`
5. Start the Figma Watchdog
    `Start-ScheduledTask -TaskName 'FigmaWatchdog'`

# FAQ
## How can I recreate the Windows Server if I need to?

```bash
terraform taint aws_instance.win2025
# Then run terraform apply as normal
```

## I cannot connect to the windows server
Add your IP address to the [infra/allowed_cidrs.txt](IP Allowlist)

