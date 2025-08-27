
# AWS Deployment

## 1. Create a Security Group

TODO

## 2. Create a Target Group

TODO

## 3. Create a Load Balancer

TODO

## 4. Create an RSA Key Pair

TODO

## 5. Create an EC2 Instance

1. Windows
1. AMI = Microsoft Windows Server 2025 Base
1. Key Pair created above
1. t3.medium
1. Create Security Group
    1. Inbound Rules
        1. RDP from your IP
        1. HTTP to port 3846 from Security Group created in `Step 1`
    1. Outbound Roles
        1. All Traffic -> 0.0.0.0/0


## 6. Install Figma and the Proxy

1. Get the Windows password
1. Connect to EC2 Instance over RDP
1. (Optional, but recommended) Change Windows password
1. Install chocolatey

```sh
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```

5. Install Git

```sh
choco install git -y
```

6. Install Golang

```sh
choco install golang -y
```

7. [Install Figma](https://www.figma.com/download/desktop/win)
    1. Log in
    1. Turn on Dev Mode MCP Server
8. Open Firewall for Port 3846

```sh
New-NetFirewallRule -DisplayName "Allow App Port 3846" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 3846
```

9. Clone the Repo

```sh
git clone https://github.com/bitovi/figma-mcp-proxy.git
```


10. Run the Proxy server

From the `figma-mcp-proxy` directory:

```sh
$env:EXTERNAL_DNS_NAME='<load balancer URL>'; & go run main.go
```


## 7.Add EC2 Instance to Target Group

To start routing traffic from the ALB to this instance:

TODO
