#!/bin/bash

set -e

make build-binary
PLUGIN_NAME="nutanix-kvdh-build"
PLUGIN_VERSION="2.3.3"
OS_ARCH="linux_amd64"
BINARY_NAME="terraform-provider-${PLUGIN_NAME}_v${PLUGIN_VERSION}"
cp bin/.terraform/plugins/registry.terraform.io/nutanix/nutanix/1.99.99/linux_amd64/terraform-provider-nutanix $BINARY_NAME

# Build your Go binary (optional if already built)
# GOOS=linux GOARCH=amd64 go build -o ${BINARY_NAME}

# Prepare directory
WORKDIR="bin/.terraform/plugins/registry.terraform.io/nutanix-kvdh/${PLUGIN_NAME}/${PLUGIN_VERSION}"
mkdir -p "${WORKDIR}/${OS_ARCH}"
cp "${BINARY_NAME}" "${WORKDIR}/${OS_ARCH}/"

# Create zip archive
cd "${WORKDIR}"
ZIP_NAME="terraform-provider-${PLUGIN_NAME}_${PLUGIN_VERSION}_${OS_ARCH}.zip"
zip -r "${ZIP_NAME}" "${OS_ARCH}"
HASH="h1:$(sha256sum "${ZIP_NAME}" | awk '{print $1}' | xxd -r -p | base64)"
cd -

# Compute hash
# Create manifest.json
cat > "${WORKDIR}/manifest.json" <<EOF
{
  "archives": {
    "${OS_ARCH}": {
      "hashes": [
        "${HASH}"
      ],
      "url": "${ZIP_NAME}"
    }
  }
}
EOF

echo "✅ Done. Your plugin mirror files are in: ${WORKDIR}"
