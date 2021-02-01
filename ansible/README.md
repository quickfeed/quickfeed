# How to Build and Deploy QuickFeed with Ansible

This document explains how to build and deploy QuickFeed for both development and production environments.

## Development

### Installing in a docker container

```ansible
% ansible-playbook -i inventory/hosts -l development dev-docker.yml
```

If you are on macOS, you probably need to provide a different python path.
Here is using homebrew python3 installation:

```ansible
% ansible-playbook -i inventory/hosts -l development dev-docker.yml -e 'ansible_python_interpreter=/usr/local/bin/python3'
```

### Installing in the local machine

```ansible
% ansible-playbook -i inventory/hosts -l development dev-local.yml
```
