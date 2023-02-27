# Faros Ingress

Faros Ingress is a minimal universal ingress controller for EVERYTHING
Using `faros ingress` you can expose any service on any port to the internet without any configuration.

# Install CLI

CLI itself is a single binary that can be used as a standalone CLI or as a
Kubectl plugin. The only difference is the name of the binary.

For `kubectl` plugin:
```bash
# Download latest CLI binary from release page
https://github.com/faroshq/faros-ingress/releases/latest

tar -xvf kubectl-faros-ingress-v*.tar.gz
mv kubectl-faros-ingress /usr/local/bin/kubectl-faros-ingress
chmod +x /usr/local/bin/kubectl-faros-ingress
rm kubectl-faros-ingress-v*.tar.gz
```

For standalone CLI:
```bash
# Download latest CLI binary from release page
https://github.com/faroshq/faros-ingress/releases/latest

tar -xvf kubectl-faros-ingress-v*.tar.gz
mv kubectl-faros-ingress /usr/local/bin/faros-ingress
chmod +x /usr/local/bin/faros-ingress
rm kubectl-faros-ingress-v*.tar.gz
```

# Usage

Usage is simple. Login to your account and create a new tunnel.

```bash
faros login
faros expose http://localhost:8080
```

More advanced usage allows one to reserve a custom domain name, specify a port
and pre-create a token for automation.

# Roadmap

* Tests!
* Add a feature to use multiple gateways for redundancy and scaling


