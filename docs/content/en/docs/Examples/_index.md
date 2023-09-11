---
title: Examples
weight: 3
date: 2017-01-05
description: See your project in action!
---

# O11y Examples

This page provides practical example configurations to help you understand how
to set up and use the O11y platform. The following example is a basic setup,
often used for initial testing and local deployments.

## Basic Localhost Configuration

In this example, we demonstrate how to set up a basic O11y deployment using
localhost as the target server. This configuration is particularly useful for
local testing or development environments.

The heart of this example lies in the `o11y.yml` configuration file. Below is
the example content:

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
      static_confis:
      - targets:
          - 'localhost:22'
        labels:
          group: 'o11y'
```

In this configuration:

* `ansible_ssh_key_path:` Specifies the path to the SSH private key used for secure
connection to the target server.
* `ansible_become_password_file:` Indicates the file containing the sudo password
for privileged operations.
* `target_groups:` Defines the group of servers that will be monitored and managed.
In this case, it includes modules for Linux, Prometheus, and Grafana.
* `targets:` Specifies the target server(s), in this case, localhost.

After creating this configuration file, running the `o11y-deploy` command will
initiate the deployment, and you'll have Prometheus and Grafana running on your
localhost.
