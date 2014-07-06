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
    "time"
    "os"
    "strconv"
    "sync"
)

func mempoolPoll(cli *btcrpcclient.Client) {
    for {
        hashes, err := cli.GetRawMempool()
        if err != nil {
            log.Fatal(err)
        }
        log.Printf("Mempool size: %d", len(hashes))
        time.Sleep(10 * time.Second)
    }
}

func main() {

    // Logging directory
	btcdHomeDir := btcutil.AppDataDir("btcd", false)
    dsFileMtx := sync.Mutex{}
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
        OnTxDoubleSpent: func(mempoolTxHash *btcwire.ShaHash, incomingTxHash *btcwire.ShaHash, isInBlock bool) {
            dsFileMtx.Lock()
            defer dsFileMtx.Unlock()
            log.Println(mempoolTxHash.String() + "," + incomingTxHash.String() + "," + strconv.FormatBool(isInBlock))

            f, err := os.OpenFile(filepath.Join(btcdHomeDir, "doublespends.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
            if err != nil {
                log.Printf("Error opening doublespend file: %v", err)
            }
            defer f.Close()

            if _, err = f.WriteString(mempoolTxHash.String() + "," + incomingTxHash.String() + "," + strconv.FormatBool(isInBlock) + "\n"); err != nil {
                log.Printf("Error writing doublespend file: %v", err)
            }
        },
	}

	// Connect to local btcd RPC server using websockets.
    log.Println("Logging double spends to " + btcdHomeDir)
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

    txHash, err := btcwire.NewShaHashFromStr("9e4cca76c166d0c165709ee656056e89fa7a493f7fb7d80e2ba4dbeef098f066")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("txhash: %v", txHash)
    rawTx, err := client.GetRawTransactionVerbose(txHash)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Found tx with hash %s, contains %d outputs.", txHash, len(rawTx.Vout))

    go mempoolPoll(client)

	// Wait until the client either shuts down gracefully (or the user
	// terminates the process with Ctrl+C).
	client.WaitForShutdown()
}
