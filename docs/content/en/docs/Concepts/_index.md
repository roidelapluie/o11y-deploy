---
title: Concepts
weight: 4
description: >
    Understanding the Building Blocks
---

Understanding the core concepts of O11y will help you deploy and manage your
observability ecosystem more effectively. This page aims to clarify the
essential terms and components you'll encounter when working with O11y.

## Deploy Server

The deploy server, also known as the Ansible bastion, serves as the central
point from which the O11y platform is deployed onto target servers. This server
should be equipped with all the dependencies required for O11y, bundled together
in what is referred to as `o11y-deps`.

## Target Server

Target servers are the machines that will be monitored and where the
observability tools will be installed. These servers are the end points that the
deploy server communicates with, to set up and manage the observability stack.

## Ansible

Ansible is the open-source automation tool used to deploy the O11y platform. It
provides an efficient way to manage configurations, deploy software, and
orchestrate more complex IT tasks between the deploy server and target servers.

## Module

A module in the O11y context is a configurable set of rules, exporters, and
dashboards. You can think of a module as a packaged component of the O11y
platform or something that needs to be monitored. Modules allow for the modular
and flexible arrangement of your observability stack.

## O11y-deps

O11y-deps is a bundle of self-contained dependencies needed to deploy the O11y
platform. This package contains Ansible, Python, and other required software. It
only needs to be installed on the deployer (or deploy server).

## Sudo

The `sudo` command-line utility is used during the installation process to grant
the installer the permissions needed to install system services and packages. It
is essential for executing privileged operations on both the deploy and target
servers.

## O11y-deploy

O11y-deploy is the specific software used to handle the installation of the O11y
platform. It requires the `o11y-deps` package to be pre-installed on the deploy
server. The `o11y-deploy` software reads from the `o11y.yml` configuration file
to understand which modules to install, and how to configure them.
