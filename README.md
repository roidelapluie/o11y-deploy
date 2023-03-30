# o11y-deploy

o11y-deploy is a command-line tool that deploys a Prometheus ecosystem on an
existing infrastructure. It leverages the
[o11y-deps-installer](https://github.com/roidelapluie/o11y-deps-installer) to manage the
required components, such as Ansible, and provides an easy-to-use interface for
configuring and deploying monitoring solutions.

## Features

⚠️ This is alpha software. Features are not stable and most of them are not
implemented.

- Deploys a Prometheus ecosystem on existing infrastructure
- Manages rules, targets, and UI
- Supports Prometheus service discovery using static configuration
- Built-in support for [ARA (ARA Records Ansible)](https://ara.recordsansible.org/)

## Installation

Before installing o11y-deploy, ensure that you have installed the necessary
dependencies using o11y-deps-installer:

```sh
./o11y-deps-installer
```

Next, clone the o11y-deploy repository and build the binary:

```
git clone https://github.com/roidelapluie/o11y-deploy.git
cd o11y-deploy
make init_roles
make build
```

## Usage

To deploy a Prometheus ecosystem using o11y-deploy, run the following command:

```
./o11y-deploy
```

The default configuration file is `o11y.yml`. You can specify a custom
configuration file using the `--config-file` flag:

```
./o11y-deploy --config-file /path/to/custom/config.yml
```

To enable the [ARA](https://ara.recordsansible.org/) webserver and view the
results of your Ansible runs, use the `--ara` flag:

```
./o11y-deploy --ara
```

For more flag options and detailed usage instructions, run:

```
./o11y-deploy --help
```

## Example Configuration

An example configuration file `o11y.yml` is provided:

```yaml
global:
  ansible_ssh_key_path: /path/to/ansible_private_key.pem
  ansible_become_password_file: /path/to/ansible_password.txt
target_groups:
  - name: servers
    modules:
      linux_module:
      prometheus_module:
    targets:
      static_configs:
      - targets:
          - 'example-host-1:22'
          - 'example-host-2:22'
        labels:
          group: 'o11y'
```

## License

o11y-deploy source code is released under the [Apache License
2.0](https://www.apache.org/licenses/LICENSE-2.0).

Please note that the release artifacts, which include the bundled dependencies
like Python, Alpine Linux, Ansible, and others, are subject to their respective
licenses. These dependencies are not covered by the Apache License 2.0 of
o11y-deploy.
