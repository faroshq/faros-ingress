# Faros

Faros is a minimal universal ingress controller for EVERYTHING
Using `faros` you can expose any service on any port to the internet without any configuration.

# Install CLI

```bash
VERSION=v0.0.6
DISTRO=linux
ARCH=amd64
curl -sL https://github.com/faroshq/faros-ingress/releases/latest/download/faros-${VERSION}-${DISTRO}-${ARCH}.tar.gz -o faros.tar.gz
tar -xvf faros.tar.gz
sudo mv faros /usr/local/bin
rm -f faros.tar.gz
```

# Roadmap

* Add connection status tracking. `connected` or `disconnected`.
* Merge API and gateway into one binary for simplicity
* Tests!
