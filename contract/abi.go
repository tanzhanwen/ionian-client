package contract

const abiIonian = `[
    {
      "inputs": [
        {
          "internalType": "bytes32",
          "name": "dataRoot",
          "type": "bytes32"
        },
        {
          "internalType": "uint256",
          "name": "sizeBytes",
          "type": "uint256"
        }
      ],
      "name": "appendLog",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "bytes",
          "name": "data",
          "type": "bytes"
        }
      ],
      "name": "appendLogWithData",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    }
  ]
`
