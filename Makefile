# Makefile for creating a tar.gz file from directories

# default target
all: roles


init_roles:
	cd ansible && git clone https://github.com/roidelapluie/ansible prometheus-community

# target for creating the tar.gz file
roles:
	./tar_directories.sh ansible/roles.tar.gz \
		ansible/prometheus-community/roles/node_exporter \
		ansible/prometheus-community/roles/prometheus

build: roles
	go build

# phony targets
.PHONY: all roles
