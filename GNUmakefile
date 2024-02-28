HOSTNAME=terraform.local
NAMESPACE=local
NAME=gotify
BINARY=terraform-provider-${NAME}
VERSION=0.0.1
OS_ARCH=linux_amd64

default: install

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Build and run the example code
.PHONY: example
example: install
	cd terraform-gotify-test && \
		rm -rf .terraform .terraform.lock.hcl terraform.tfstate && \
		terraform init && \
		TF_LOG=INFO terraform apply -auto-approve && \
		sleep 2 && \
		TF_LOGO=INFO terraform destroy -auto-approve

.PHONE: apply
apply: install
	cd terraform-gotify-test && \
		rm -rf .terraform .terraform.lock.hcl terraform.tfstate && \
		terraform init && \
		TF_LOG=INFO terraform apply -auto-approve

.PHONE: destroy
destroy: install
	cd terraform-gotify-test && \
		rm -rf .terraform .terraform.lock.hcl && \
		terraform init && \
		TF_LOG=INFO terraform destroy -auto-approve