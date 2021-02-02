# How to Build and Deploy QuickFeed with Ansible

This document explains how to build and deploy QuickFeed for both development and production environments.


## Development

### Installing in a docker container
```ansible
% ansible-playbook -i inventory/hosts -l development dev-docker.yml
```

### Installing in the local machine
```ansible
% ansible-playbook -i inventory/hosts -l development install.yml
```