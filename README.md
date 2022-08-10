# Ionian Client
Go implementation for client to interact with storage nodes in Ionian network.

# CLI
Run `go build` under the root folder to compile the executable binary.

**Global options**
```
  -h, --help               help for ionian-client
      --log-force-color    Force to output colorful logs
      --log-level string   Log level (default "info")
```

**Deploy contract**

```
./ionian-client deploy --url <blockchain_rpc_endpoint> --key <private_key> --bytecode <bytecode_hex_or_json_file>
```

**Upload file**
```
./ionian-client upload --url <blockchain_rpc_endpoint> --contract <ionian_contract_address> --key <private_key> --node <storage_node_rpc_endpoint> --file <file_path>
```

**Download file**
```
./ionian-client download --node <storage_node_rpc_endpoint> --root <file_root_hash> --file <output_file_path>
```
