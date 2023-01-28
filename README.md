# Faros

Faros is a minimal universal ingress controller for EVERYTHING
Using `faros` you can expose any service on any port to the internet without any configuration.

# Install CLI

```bash
# Download latest CLI binary from release page
https://github.com/faroshq/faros-ingress/releases/latest

tar -xvf faros-v*.tar.gz
mv faros /usr/local/bin/faros
chmod +x /usr/local/bin/faros
rm faros-v*.tar.gz
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

* Add connection status tracking. `connected` or `disconnected`.
* Merge API and gateway into one binary for simplicity
* Tests!
* Add a feature to use multiple gateways for redundancy and scaling


