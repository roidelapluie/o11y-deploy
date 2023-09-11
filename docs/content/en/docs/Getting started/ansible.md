---
title: Ansible user
date: 2017-01-05
description: >
  How to setup the ansible user
categories: [Examples]
tags: [test, sample, docs]
---

The Ansible user plays a crucial role in the O11y deployment process. This user
serves as the account under which Ansible operations are executed on the target
servers. In this guide, we'll delve into how to configure the Ansible user in
the `o11y.yml` file, as well as how to set up this user on your Unix-based target
machines.

## Configuring Ansible User in the configuration file

In your o11y.yml configuration file, you have several options to fine-tune the
behavior of the Ansible user:

* `ansible_ssh_key_path`: The path to the SSH private key that Ansible will use to
  log into the target servers.
* `ansible_become_password_file`: The file containing the password for escalating to
  sudo privileges.
* `ansible_user`: Specifies the username that Ansible will use to connect to the
  target machines (Default: `ansible`).
* `ansible_trust_on_first_use`: Whether or not to trust the target machine upon the
first SSH connection (Default: `true`).

Example snippet from `o11y.yml`:

```yaml
global:
  ansible_ssh_key_path: /path/to/ssh_private_key.pem
  ansible_become_password_file: /path/to/sudo_password.txt
  ansible_user: ansible
  ansible_trust_on_first_use: true
```

## Setting Up the Ansible User on Target Machines

Below are the Unix commands to set up the Ansible user on the target machine,
generate an SSH key on the deployer machine, copy the key to the target machine,
change the password, and grant sudo privileges.

### Create the User on Target Machine

SSH into your target machine:

```
ssh username@target_machine_ip
```

Create the Ansible user:

```
sudo adduser ansible
```

Change User Password on Target Machine

```
sudo passwd ansible
```

Save this password in a text file on the deployer machine, for example
`/path/to/sudo_password.txt`.

### Generate SSH Key on Deployer Machine

On your deployer machine, generate an SSH key pair:

```
ssh-keygen -t rsa -f /path/to/ssh_private_key.pem
```

### Copy SSH Key to Target Machine

Use `ssh-copy-id` to copy the public key to your target machine:

```
ssh-copy-id -i /path/to/ssh_private_key.pem.pub ansible@target_machine_ip
```

Alternatively, if you can's SSH  with a password, copy the file manually, as
follow.

First, on your deployer machine, display the contents of your SSH public key
using the cat command:

```
cat /path/to/ssh_private_key.pem.pub
```

This will display the public key in the terminal. Copy the entire key string to
your clipboard.

SSH into your target machine using a method that is allowed (perhaps using an
existing key or other authentication method):

```
ssh username@target_machine_ip
```

Once logged into the target machine, switch to the Ansible user or stay as the
current user if you have the necessary permissions:

```
su - ansible
```

Or remain as the current user if you have permissions to edit the Ansible user's
home directory.

Navigate to the `.ssh` directory in the Ansible user's home directory, creating it
if it doesn't exist:

```
mkdir -p ~/.ssh
cd ~/.ssh
```

Open the `authorized_keys` file using a text editor like vim or nano, or create it
if it doesn't exist:

```
touch authorized_keys
nano authorized_keys
```

Paste the copied SSH public key string into a new line in this file, then save
and close the file.

Finally, set the proper permissions for the `.ssh` directory and the
`authorized_keys` file:

```
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

### Change User Password on Target Machine


Open the sudoers file using `visudo`:

```
sudo visudo
```

Add the following line to grant sudo access to the Ansible user:

```
ansible ALL=(ALL) NOPASSWD: ALL
```

## Summary

You've successfully configured the Ansible user both within the o11y.yml file
and on your target machine. This user is now equipped with the permissions and
keys necessary to facilitate the O11y deployment.
