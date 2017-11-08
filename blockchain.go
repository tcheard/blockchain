package main

import (
	"encoding/hex"
	"errors"
	"os"

	"github.com/boltdb/bolt"
	perrors "github.com/pkg/errors"
)

const (
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

// Blockchain represents the actual blockchain holding all its blocks
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// NewBlockchain creates a new blockchain by reading from the database
func NewBlockchain() (*Blockchain, error) {
	if !dbExists() {
		return nil, errors.New("create a blockchain first")
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, perrors.Wrap(err, "failed to open database file")
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		return nil, perrors.Wrap(err, "failed to update database")
	}

	return &Blockchain{tip: tip, db: db}, nil
}

// CreateBlockchain starts a brand new blockchain
func CreateBlockchain(address string) (*Blockchain, error) {
	if dbExists() {
		return nil, errors.New("blockchain already exists")
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, perrors.Wrap(err, "failed to open database file")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx, err := NewCoinbaseTransaction(address, genesisCoinbaseData)
		if err != nil {
			return perrors.Wrap(err, "failed to create new coinbase transaction")
		}

		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return perrors.Wrap(err, "failed to create database bucket")
		}

		genSer, err := genesis.Serialize()
		if err != nil {
			return perrors.Wrap(err, "failed to serialize genesis block")
		}

		err = b.Put(genesis.Hash, genSer)
		if err != nil {
			return perrors.Wrap(err, "failed to put the genesis block")
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			return perrors.Wrap(err, "failed to put the 'l' key")
		}

		tip = genesis.Hash

		return nil
	})

	if err != nil {
		return nil, perrors.Wrap(err, "failed to update database")
	}

	return &Blockchain{tip: tip, db: db}, nil
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0

	unspentTXs, err := bc.FindUnspentTransactions(pubKeyHash)
	if err != nil {
		return 0, nil, perrors.Wrap(err, "failed to retrieve unspent transactions")
	}

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs, nil
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs for a given address
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) ([]Transaction, error) {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		b, err := bci.Next()
		if err != nil {
			return nil, perrors.Wrap(err, "failed to retrieve next block")
		}

		for _, tx := range b.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					usesKey, err := in.UsesKey(pubKeyHash)
					if err != nil {
						return nil, perrors.Wrap(err, "failed to determine if tx.in uses key")
					}

					if usesKey {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(b.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs, nil
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) ([]*TXOutput, error) {
	var UTXOs []*TXOutput

	unspentTXs, err := bc.FindUnspentTransactions(pubKeyHash)
	if err != nil {
		return nil, perrors.Wrap(err, "failed to find unspent transactions")
	}

	for _, tx := range unspentTXs {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs, nil
}

// Iterator retrieves an iterator for the blockchain
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		currentHash: bc.tip,
		db:          bc.db,
	}
}

// MineBlock creates a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) error {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return perrors.Wrap(err, "failed to retrieve last hash")
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		nbSer, err := newBlock.Serialize()
		if err != nil {
			return perrors.Wrap(err, "failed to serialize new block")
		}

		err = b.Put(newBlock.Hash, nbSer)
		if err != nil {
			return perrors.Wrap(err, "failed to put the new block")
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return perrors.Wrap(err, "failed to put the 'l' key")
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		return perrors.Wrap(err, "failed to write new block")
	}

	return nil
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
