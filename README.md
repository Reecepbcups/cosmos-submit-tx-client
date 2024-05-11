# cosmos-submit-tx-client

execute `go run *.go` to run. Make sure to change the source to use your own mnemonic, chain-id, and RPC or gRPC endpoint.

## Example

```
(0) reece@reece-desktop10 [~/Desktop/go-cosmos-client]
(main) -> $ go run *.go

txBytesJson {"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"cosmos165smcg7d2fqtwewj3uf33rc33dh8h46yns3sm5","to_address":"cosmos1vetzf8f4jnfkdpynytx2znycc6sewcem8flcnm","amount":[{"denom":"uatom","amount":"1"}]}],"memo":"my test memo","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A6Y2GkP4FFa/x0tK2rORnQjnUgmklhmctM3Zd6pqSO39"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"11"}],"fee":{"amount":[{"denom":"uatom","amount":"750"}],"gas_limit":"100000","payer":"","granter":""},"tip":null},"signatures":["dCFTglv0rfgQ12HBeOhpfYf+D39+YMqTAYjN2AZ/j5l6H0Tcyzaz8w//uxC9DFsY+Z3Khk2zhqjVjaAnSv6PwA=="]}

broadcastRes: tx_response:{txhash:"A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3"  raw_log:"[]"}
POLLING: error getting tx. Trying again...  rpc error: code = NotFound desc = tx not found: A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3
POLLING: error getting tx. Trying again...  rpc error: code = NotFound desc = tx not found: A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3
POLLING: error getting tx. Trying again...  rpc error: code = NotFound desc = tx not found: A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3
POLLING: error getting tx. Trying again...  rpc error: code = NotFound desc = tx not found: A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3
POLLING: error getting tx. Trying again...  rpc error: code = NotFound desc = tx not found: A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3
POLLING: error getting tx. Trying again...  rpc error: code = NotFound desc = tx not found: A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3
POLLING: error getting tx. Trying again...  rpc error: code = NotFound desc = tx not found: A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3
POLLING: error getting tx. Trying again...  rpc error: code = NotFound desc = tx not found: A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3

txRes: code: 0
codespace: ""
data: 0A1E0A1C2F636F736D6F732E62616E6B2E763162657461312E4D736753656E64
events: []
gas_used: "82568"
gas_wanted: "100000"
height: "18711936"
info: ""
logs: []
raw_log: '[{"events":[{"type":"coin_received","attributes":[{"key":"receiver","value":"cosmos1vetzf8f4jnfkdpynytx2znycc6sewcem8flcnm"},{"key":"amount","value":"1uatom"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"cosmos165smcg7d2fqtwewj3uf33rc33dh8h46yns3sm5"},{"key":"amount","value":"1uatom"}]},{"type":"message","attributes":[{"key":"action","value":"/cosmos.bank.v1beta1.MsgSend"},{"key":"sender","value":"cosmos165smcg7d2fqtwewj3uf33rc33dh8h46yns3sm5"},{"key":"module","value":"bank"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"cosmos1vetzf8f4jnfkdpynytx2znycc6sewcem8flcnm"},{"key":"sender","value":"cosmos165smcg7d2fqtwewj3uf33rc33dh8h46yns3sm5"},{"key":"amount","value":"1uatom"}]}]}]'
timestamp: "2024-01-14T03:03:20Z"
tx: null
txhash: A71C62A2043E4050129CDE12C0D245DA3A7C8C6A590CC588D423EB2EB59304F3
```