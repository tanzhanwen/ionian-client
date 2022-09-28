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
  ]`

const abiFlow = `[
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "sender",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "uint256",
        "name": "index",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "bytes32",
        "name": "startMerkleRoot",
        "type": "bytes32"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "submissionIndex",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "flowLength",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "bytes32",
        "name": "context",
        "type": "bytes32"
      }
    ],
    "name": "NewEpoch",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "sender",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "bytes32",
        "name": "identity",
        "type": "bytes32"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "submissionIndex",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "startPos",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "length",
        "type": "uint256"
      },
      {
        "components": [
          {
            "internalType": "uint256",
            "name": "length",
            "type": "uint256"
          },
          {
            "components": [
              {
                "internalType": "bytes32",
                "name": "root",
                "type": "bytes32"
              },
              {
                "internalType": "uint256",
                "name": "height",
                "type": "uint256"
              }
            ],
            "internalType": "struct IonianSubmissionNode[]",
            "name": "nodes",
            "type": "tuple[]"
          }
        ],
        "indexed": false,
        "internalType": "struct IonianSubmission",
        "name": "submission",
        "type": "tuple"
      }
    ],
    "name": "Submission",
    "type": "event"
  },
  {
    "inputs": [],
    "name": "getContext",
    "outputs": [
      {
        "components": [
          {
            "internalType": "uint256",
            "name": "epoch",
            "type": "uint256"
          },
          {
            "internalType": "uint256",
            "name": "epochStart",
            "type": "uint256"
          },
          {
            "internalType": "bytes32",
            "name": "flowRoot",
            "type": "bytes32"
          },
          {
            "internalType": "uint256",
            "name": "flowLength",
            "type": "uint256"
          },
          {
            "internalType": "bytes32",
            "name": "digest",
            "type": "bytes32"
          }
        ],
        "internalType": "struct MineContext",
        "name": "",
        "type": "tuple"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "digest",
        "type": "bytes32"
      }
    ],
    "name": "getEpochRange",
    "outputs": [
      {
        "components": [
          {
            "internalType": "uint128",
            "name": "start",
            "type": "uint128"
          },
          {
            "internalType": "uint128",
            "name": "end",
            "type": "uint128"
          }
        ],
        "internalType": "struct EpochRange",
        "name": "",
        "type": "tuple"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "makeContext",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "numSubmissions",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
        {
          "components": [
            {
              "internalType": "uint256",
              "name": "length",
              "type": "uint256"
            },
            {
              "internalType": "bytes",
              "name": "tags",
              "type": "bytes"
            },
            {
              "components": [
                {
                  "internalType": "bytes32",
                  "name": "root",
                  "type": "bytes32"
                },
                {
                  "internalType": "uint256",
                  "name": "height",
                  "type": "uint256"
                }
              ],
              "internalType": "struct IonianSubmissionNode[]",
              "name": "nodes",
              "type": "tuple[]"
            }
          ],
          "internalType": "struct IonianSubmission",
          "name": "submission",
          "type": "tuple"
        }
      ],
    "name": "submit",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      },
      {
        "internalType": "bytes32",
        "name": "",
        "type": "bytes32"
      },
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]`
