# Shutter API Documentation for dApp Developers

Welcome to the **Shutter API** documentation! This guide will help you integrate Shutter's Commit and Reveal Scheme into your decentralized application (dApp). The Shutter system provides a secure, decentralized, and tamper-proof commit-and-reveal workflow, ensuring integrity and confidentiality in your application.

---
## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Endpoints](#endpoints)
  - [Register an Identity with a Decryption Trigger](#1-register-an-identity-with-a-decryption-trigger)
  - [Retrieve the Encryption Data](#2-retrieve-the-encryption-data)
  - [Retrieve the Decryption Key](#3-retrieve-the-decryption-key)
  - [Decrypt Commitments](#4-decrypt-commitments)
4. [Future features](#future-features)
5. [FAQs](#faqs)
6. [Swagger Documentation](#swagger-documentation)
7. [Support](#support)

---

## Overview

The Shutter system leverages threshold encryption, distributed cryptographic operations, and a decentralized infrastructure to handle commitments securely. Core components include:

- **Registry Contract**: An on-chain contract where clients register identities and specify time-based decryption triggers.
- **Keypers**: A distributed set of nodes that monitor the registry contract, handle cryptographic operations such as distributed key generation, and release decryption keys securely.
- **API**: An API that simplifies interaction with the Shutter system by exposing endpoints for encryption and decryption operations.

This documentation will guide you through:
- Setting up identities and time-based decryption triggers.
- Retrieving encryption data and decryption keys.
- Decrypting encrypted commitments.

---

## Prerequisites

- **API Access to the Shutter Centralized Service**:  
  At the moment, the access is free of charge. You only need to query the API endpoints at the addresses below:
  - **Chiado**: `http://64.227.118.171:8001/api/[ADD_ENDPOINT]`
  - **Mainnet**: TBD

- **Address of the Shutter Registry Contract**:
  - **Chiado Address**: `0x43D1Aee2D61fb206b72c6bDd8a0F17Eb6BF1eF51`
  - **Gnosis Address**: TBD

- **Address of the API**:
  - **Chiado Address**: TBD
  - **Gnosis Address**: TBD
---

## Endpoints

### 1. Register an Identity with a Decryption Trigger

To begin using the Shutter system, register an identity and specify a time-based decryption trigger. This step links an identity to a decryption key and sets the release conditions for the key to a Unix timestamp.

Refer to the `/register_identity` endpoint in the Swagger documentation for details on parameters and responses.

> **Note**: When registering identities through our API, the API account address is used to compute the identity that will be returned. If you want to use your own address, you need to submit the registration directly to the registry contract. The contract's definition can be found here:  
> [ShutterRegistry.sol](https://github.com/shutter-network/contracts/blob/main/src/shutter-service/ShutterRegistry.sol#L1C1-L86C2).

#### Example Request
```bash
curl -X POST http://<API_BASE_URL>/register_identity \
-H "Content-Type: application/json" \
-d '{
  "decryptionTimestamp": 1735044061,
  "identityPrefix": "0x79bc8f6b4fcb02c651d6a702b7ad965c7fca19e94a9646d21ae90c8b54c030a0"
}'
```

#### Example Response
```json
{
  "eon": 1,
  "eon_key": "0x57af5437a84ef50e5ed75772c18ae38b168bb07c50cadb65fc6136604e662255",
  "identity": "0x8c232eae4f957259e9d6b68301d529e9851b8642874c8f59d2bd0fb84a570c75",
  "identity_prefix": "0x79bc8f6b4fcb02c651d6a702b7ad965c7fca19e94a9646d21ae90c8b54c030a0",
  "tx_hash": "0x3026ad202ca611551377eef069fb6ed894eae65329ce73c56f300129694f12ba"
}
```

### 2. Retrieve the Encryption Data

To encrypt commitments, obtain the encryption data associated with your identity. Use the `/get_data_for_encryption` endpoint to retrieve all necessary encryption data.

Refer to the Swagger documentation for specifics on this endpoint.

#### Example Request
```bash
curl -X GET "http://<API_BASE_URL>/get_data_for_encryption?address=0xb9C303443c9af84777e60D5C987AbF0c43844918&identityPrefix=0x79bc8f6b4fcb02c651d6a702b7ad965c7fca19e94a9646d21ae90c8b54c030a0"
```

#### Example Response
```json
{
"eon": 1,
"eon_key": "0x57af5437a84ef50e5ed75772c18ae38b168bb07c50cadb65fc6136604e662255",
"identity": "0x8c232eae4f957259e9d6b68301d529e9851b8642874c8f59d2bd0fb84a570c75",
"identity_prefix": "0x79bc8f6b4fcb02c651d6a702b7ad965c7fca19e94a9646d21ae90c8b54c030a0"
}
```

#### Encrypting Commitments in Go

The following Go code demonstrates how to use the encryption data retrieved from the Shutter API to encrypt commitments:

```go
// NOTE: This example requires the "github.com/shutter-network/shutter/shlib/shcrypto" package.
// Make sure to install it in your Go environment before running this code.
package main

import (
  "crypto/rand"
  "encoding/hex"
  "fmt"
  "log"
  "strings"

  "github.com/shutter-network/shutter/shlib/shcrypto"
)

func main() {
  // Encryption data provided by the Shutter API
  identityHex := "0x8c232eae4f957259e9d6b68301d529e9851b8642874c8f59d2bd0fb84a570c75"
  eonPublicKeyHex := "0x57af5437a84ef50e5ed75772c18ae38b168bb07c50cadb65fc6136604e662255"
  message := []byte("please hide this message")

  identityHex = strings.TrimPrefix(identityHex, "0x")
  eonPublicKeyHex = strings.TrimPrefix(eonPublicKeyHex, "0x")

  // Convert hex strings to bytes
  identity, err := hex.DecodeString(identityHex)
  if err != nil {
    log.Fatalf("Failed to decode identity: %v", err)
  }

  eonPublicKeyBytes, err := hex.DecodeString(eonPublicKeyHex)
  if err != nil {
    log.Fatalf("Failed to decode eon public key: %v", err)
  }

  // Create EonPublicKey struct from bytes
  eonPublicKey := &shcrypto.EonPublicKey{}
  if err := eonPublicKey.Unmarshal(eonPublicKeyBytes); err != nil {
    log.Fatalf("Failed to unmarshal EonPublicKey: %v", err)
  }

  // Compute the Epoch ID from the identity
  epochID := shcrypto.ComputeEpochID(identity)

  // Generate a random sigma
  sigma, err := shcrypto.RandomSigma(rand.Reader)
  if err != nil {
    log.Fatalf("Failed to generate random sigma: %v", err)
  }

  // Encrypt the message
  encryptedCommitment := shcrypto.Encrypt(message, eonPublicKey, epochID, sigma)

  // Marshal the encrypted commitment into bytes
  encryptedCommitmentBytes := encryptedCommitment.Marshal()

  // Convert to hex string
  encryptedCommitmentHex := "0x" + hex.EncodeToString(encryptedCommitmentBytes)

  // Print the encrypted commitment
  fmt.Printf("Encrypted Commitment: %s\n", encryptedCommitmentHex)
}
```

You can also translate the code into JavaScript if you prefer to use the JavaScript SDK:
[Shutter JavaScript SDK on npm](https://www.npmjs.com/package/@shutter-network/shutter-crypto)

### 3. Retrieve the Decryption Key

After the decryption trigger conditions are met (i.e., the specified timestamp has passed), retrieve the decryption key using the `/get_decryption_key` endpoint.

Refer to the Swagger documentation for detailed usage.

#### Example Request
```bash
curl -X GET "http://<API_BASE_URL>/get_decryption_key?identity=0x8c232eae4f957259e9d6b68301d529e9851b8642874c8f59d2bd0fb84a570c75"
```

#### Example Response
```json
{
  "decryption_key": "0x99a805fc26812c13041126b25e91eccf3de464d1df7a95d1edca8831a9ec02dd",
  "decryption_timestamp": 1735044061,
  "identity": "0x8c232eae4f957259e9d6b68301d529e9851b8642874c8f59d2bd0fb84a570c75"
}
```

### 4. Decrypt Commitments

Once you have the decryption key, use it to decrypt commitments encrypted with the Shutter system. The `/decrypt_commitment` endpoint enables this process.

Refer to the Swagger documentation for endpoint details.

#### Example Request
```bash
curl -X GET "http://<API_BASE_URL>/decrypt_commitment?identity=0x8c232eae4f957259e9d6b68301d529e9851b8642874c8f59d2bd0fb84a570c75&encryptedCommitment=0x03b5685a460a95ba628e04b24155d6722f7c4e376a1627f714a4ae9cecd2982e005eff12ac8150b8842c29f8d5eaf4d0da0b626f762b4826d779d8969b577acb28df96cab026aa57c00cd74b07ca51e8c0c1a59933e29a728311900ebfc26c6804260914c96cb10dbd6d2ed3f6cb77788a74b5aae5f4ce6f40be53310a0524d42d5a6f03b5c1517ec097553733e228276fcdfc4b569f7ef4311a461d68819d634c"
```

#### Example Response
```json
{
  "decrypted_message": "please hide this message"
}
```

> **Note**: Replace `<API_BASE_URL>` in all example requests with the actual base URL for the API, found in the pre-requisite section, such as `http://64.227.118.171:8001/api`.

## Future Features

- **Event-Based and Block-Based Triggers**  
  Future versions of the Shutter system will support event-based and block-based decryption triggers for enhanced functionality.

- **Real-Time Notifications**  
  Planned updates include WebSocket-based notifications for real-time key releases, improving user experience and interactivity.

## FAQs

### What happens if a keyper experiences downtime?
The keyper set is designed to handle downtime gracefully. Any missed decryption key releases will be sent upon recovery.

### How secure is the Shutter system?
The Shutter system uses threshold encryption and distributed cryptographic operations to ensure that no single entity can compromise the security of commitments.

## Swagger Documentation

For detailed API specifications, including parameters, responses, and error codes, visit the Swagger Documentation:

- [Chiado Swagger Documentation](http://64.227.118.171:8001/docs/index.html)
- **Mainnet**: Documentation link TBD

## Support

For additional support or inquiries:
- Contact the Shutter development team.
- Open an issue on our GitHub repository.

---

Thank you for using Shutter! Together, we are building a more secure and decentralized future.