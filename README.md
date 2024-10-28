# Textfully Go SDK

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/gtfol/textfully-go.svg)](https://pkg.go.dev/github.com/gtfol/textfully-go)
![Release](https://img.shields.io/github/release/gtfol/textfully-go.svg?style=flat-square)
---

The official Go SDK for [Textfully](https://textfully.dev) — The Open Source Twilio Alternative.

## Installation

```bash
go get github.com/gtfol/textfully-go/v1
```

## Setup

First, you need to generate an API key from the [Textfully Dashboard](https://textfully.dev/dashboard/api/keys).

## Quick Start

```go
import (
    "log"
    "github.com/gtfol/textfully-go/v1"
)

func main() {
    // Set your API key
    client := textfully.New("tx_123")

    // Send a message
    _, err := client.Send(
        "+16175555555", // verified phone number
        "Hello, world!",
    )
    if err != nil {
        log.Fatal(err)
    }
}

```
