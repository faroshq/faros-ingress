# Faros

Faros is a minimal universal ingress controller for Everything
Using faros you can expose any service on any port to the internet


# Install CLI:

```bash
# TODO: once public repo, this should work
VERSION=v0.0.1
DISTRO=linux
ARCH=amd64
curl -sL https://github.com/mjudeikis/portal/releases/latest/download/faros-${VERSION}-${DISTRO}-${ARCH}.tar.gz -o faros.tar.gz
tar -xvf faros.tar.gz
sudo mv faros /usr/local/bin
rm -f faros.tar.gz
```
