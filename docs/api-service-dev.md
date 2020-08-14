## API service developer guide

# Steps :

- Developers can use a proto lint extension while making changes to the proto file.
eg: if you're using VS code, use [vscode-proto3](https://marketplace.visualstudio.com/items?itemName=zxh404.vscode-proto3).

- After any change in the [proto file](../ndm.proto), the corresponding `golang` code should be generated. To generate run: 

```make protos```

- The generated code can be found under [spec/ndm](../spec/ndm/ndm.pb.go)

- The code related new functionality for the API service should be added under the `api-service` directory. Based on the type of functionality added, it should be packaged as follows:
    - cluster: all services that are common to the cluster, like version of NDM.
    - node: services that are node dependent, like list of blockdevices, checking iSCSI status etc.
