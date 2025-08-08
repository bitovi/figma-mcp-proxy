
# Useful OS commands

## Firewall

### Windows

Allow port 3846 through Windows Firewall:

```sh
New-NetFirewallRule -DisplayName "Allow Port 3846" -Direction Inbound -Protocol TCP -LocalPort 3846 -Action Allow -Profile Any
```

## Open Figma

### MacOS

```sh
open "figma://"
```

### Windows

```sh
Start-Process "figma://"
```

## Log in to Figma

### MacOS

```sh
TODO
```

### Windows

```sh
TODO
```

## Open Figma Design Document

### MacOS

```sh
open "figma://design/{file_key}/{file_name}"
```

### Windows

```sh
Start-Process "figma://design/{file_key}/{file_name}"
```
