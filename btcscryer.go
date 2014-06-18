// Copyright (c) 2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/conformal/btcjson"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"io/ioutil"
	"log"
	"path/filepath"
)

func main() {
	// Only override the handlers for notifications you care about.
	// Also note most of these handlers will only be called if you register
	// for notifications.  See the documentation of the btcrpcclient
	// NotificationHandlers type for more details about each handler.
	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnBlockConnected: func(hash *btcwire.ShaHash, height int32) {
			log.Printf("Block connected: %v (%d)", hash, height)
		},
		OnBlockDisconnected: func(hash *btcwire.ShaHash, height int32) {
			log.Printf("Block disconnected: %v (%d)", hash, height)
		},
        OnTxAcceptedVerbose: func(txDetails *btcjson.TxRawResult) {
            log.Printf("Tx Mempooled: %s", txDetails.Txid)
        },
	}

	// Connect to local btcd RPC server using websockets.
	btcdHomeDir := btcutil.AppDataDir("btcd", false)
	certs, err := ioutil.ReadFile(filepath.Join(btcdHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}
	connCfg := &btcrpcclient.ConnConfig{
		Host:         "localhost:8334",
		Endpoint:     "ws",
		User:         "yourrpcuser",
		Pass:         "yourrpcpass",
		Certificates: certs,
	}
	client, err := btcrpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Fatal(err)
	}

	// Register for block connect and disconnect notifications.
	if err := client.NotifyBlocks(); err != nil {
		log.Fatal(err)
	}
	log.Println("NotifyBlocks: Registration Complete")

    if err := client.NotifyNewTransactions(true); err != nil {
        log.Fatal(err)
    }
	log.Println("NotifyNewTransactions: Registration Complete")

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)

	// Wait until the client either shuts down gracefully (or the user
	// terminates the process with Ctrl+C).
	client.WaitForShutdown()
}
