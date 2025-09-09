provider "aws" {
  region = var.aws_region
}

locals {
    short_environment = substr(var.target_environment, 0, 1)
    // App name = figma-bridge, short_app_name should be the first letter of each word
    short_app_name = join("", [for word in split("-", var.app_name) : substr(word, 0, 1)])
    fully_qualified_name = "${local.short_app_name}-${var.client_name}-${local.short_environment}"
    private_key_pem = var.private_key_path != "" ? file(var.private_key_path) : ""
    allowed_cidrs_raw = var.allowed_cidrs_file != "" && fileexists(var.allowed_cidrs_file) ? file(var.allowed_cidrs_file) : ""
    allowed_cidrs = [
        for l in split("\n", local.allowed_cidrs_raw) : trimspace(l)
        if trimspace(l) != "" && !startswith(trimspace(l), "#")
    ]
}