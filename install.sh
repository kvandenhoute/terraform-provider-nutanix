#!/bin/bash

version="1.99.99"
provider="nutanix"
provider_binary="terraform-provider-$provider"

# Function to create terraform.rc file
create_terraform_rc() {
    cat <<EOF > ~/.terraformrc
provider_installation {
  filesystem_mirror {
    path = "$PWD/bin/.terraform/plugins"
  }
  direct {
    exclude = ["registry.terraform.io/*/*"]
  }
}
EOF
    echo "Created terraform.rc at ~/.terraformrc"
}

# Function to create provider.tf file
create_provider_config() {
    cat <<EOF > provider.tf
terraform {
  required_providers {
    $provider = {
      source = "nutanix/$provider"
      version = "$version"
    }
  }
}
EOF
    echo "Created provider.tf configuration"
}

# Main setup
if [ ! -f "bin/$provider_binary" ]; then
    echo "Error: Provider binary not found in current directory: $provider_binary" >&2
    exit 1
fi

# Create directory structure
provider_dir="bin/.terraform/plugins/registry.terraform.io/nutanix/$provider/$version/linux_amd64"
mkdir -p "$provider_dir"
echo "Created provider directory structure"

# Copy provider binary
cp "bin/$provider_binary" "$provider_dir/"
echo "Copied provider binary to: $provider_dir/$provider_binary"

# Create terraform.rc file
create_terraform_rc

# Create provider configuration
create_provider_config

echo -e "\nSetup completed successfully!\n"
echo "Next steps:"
echo "1. Run twtrito initialize your working directory"
echo "2. Create your terraform configuration files (.tf)"
echo "3. Run 'terraform plan' to verify the setup"