# Makefile for creating a tar.gz file from directories

# default target
all: roles


init_roles:
	cd ansible \
	    && (git clone https://github.com/roidelapluie/ansible prometheus-community || (cd prometheus-community && git pull) ) \
		&& (git clone https://github.com/grafana/grafana-ansible-collection || (cd grafana-ansible-collection && git pull)) \
		&& cd grafana-ansible-collection && git reset --hard 2.2.3

# target for creating the tar.gz file
roles:
	./tar_directories.sh ansible/roles.tar.gz \
		ansible/prometheus-community/roles/alertmanager \
		ansible/prometheus-community/roles/node_exporter \
		ansible/prometheus-community/roles/prometheus \
		ansible/grafana-ansible-collection/roles/grafana \
		ansible/roles/authp

build-frontend:
	cd frontend/ui && npm install && npm run build
	cd frontend/ui && ./compress_assets.sh
	cd frontend && go build -tags builtinassets -o o11y-deploy-frontend

build: roles
	go build

# phony targets
.PHONY: all roles
