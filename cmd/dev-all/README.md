   # certificate

   ```bash
   go run ./hack/genkey proxy
   mv proxy.* dev

   go run ./hack/genkey -client proxy-client
   mv proxy-client.* dev

   go run ./hack/genkey -client server
   mv server.* dev
   ```
