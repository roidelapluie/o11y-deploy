---
title: Getting Started
description: Your Guide to Effortless Observability
categories: []
tags: [test, docs]
weight: 2
---

Welcome to the Getting Started guide for O11y! Whether you're new to
observability or a seasoned veteran, this guide will walk you through the
process of setting up your own O11y platform. With O11y, you can easily deploy a
comprehensive observability ecosystem featuring Prometheus metrics, dashboards,
and alerting rules right on your on-premises infrastructure.

## Prerequisites

Before you proceed, please ensure you meet the following requirements: 

**Deployer Node**: A server where the O11y deployer can run.
**Target Servers:** The servers where the O11y platform will be installed.

Target servers should have:

* A `ansible` user with sudo access.
* SSH accessibility from the deployer node.
* SSH key for secure communication.

## Installation

### Step 1: Install Dependencies

First, download the `o11y-deps-installer` which contains all the necessary
dependencies. It can be downloaded from [here](https://github.com/roidelapluie/o11y-deps-installer).

Run the installer with the following command:

```bash
./o11y-deps-installer
```

### Step 2: Create Configuration File

Next, create a configuration file for `o11y-deploy`. Name it `o11y.yml`. Below is an
example configuration for localhost deployment:

```yaml
global:
  ansible_ssh_key_path: /path/to/ansible_private_key.pem
  ansible_become_password_file: /path/to/ansible_password.txt
target_groups:
  - name: servers
    modules:
      linux_module:
      prometheus_module:
      grafana_module:
    targets:
      static_configs:
      - targets:
          - 'localhost:22'
        labels:
          group: 'o11y'
```


### Step 3: Run the O11y Deployer

Download the `o11y-deploy` binary from [here](https://github.com/roidelapluie/o11y-deploy).

Run the deployer with:

```bash
./o11y-deploy
```

### Try It Out

Congratulations! You've successfully deployed your O11y platform.

* **Prometheus:** You can access the Prometheus server at `127.0.0.1:9090`. It is
configured to monitor Linux metrics.
* **Grafana:** You can also access Grafana for visual dashboards at `127.0.0.1:3000`.


