package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"os"

	"github.com/boltdb/bolt"
)

const (
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

// Blockchain represents the actual blockchain holding all its blocks
type Blockchain struct {
	tip []byte
	DB  *bolt.DB
}

// NewBlockchain creates a new blockchain by reading from the database
func NewBlockchain() (*Blockchain, error) {
	if !dbExists() {
		return nil, errors.New("create a blockchain first")
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Blockchain{tip: tip, DB: db}, nil
}

// CreateBlockchain starts a brand new blockchain
func CreateBlockchain(address string) (*Blockchain, error) {
	if dbExists() {
		return nil, errors.New("blockchain already exists")
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx, err := NewCoinbaseTransaction(address, genesisCoinbaseData)
		if err != nil {
			return err
		}

		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}

		genSer, err := genesis.Serialize()
		if err != nil {
			return err
		}

		err = b.Put(genesis.Hash, genSer)
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			return err
		}

		tip = genesis.Hash

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Blockchain{tip: tip, DB: db}, nil
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		b, err := bci.Next()
		if err != nil {
			return Transaction{}, err
		}

		for _, tx := range b.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(b.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("transaction is not found")
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() (map[string]*TXOutputs, error) {
	UTXO := make(map[string]*TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		b, err := bci.Next()
		if err != nil {
			return nil, err
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

				outs := UTXO[txID]
				if outs == nil {
					outs = &TXOutputs{}
				}
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(b.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO, nil
}

// Iterator retrieves an iterator for the blockchain
func (bc *Blockchain) Iterator() *BIterator {
	return &BIterator{
		currentHash: bc.tip,
		db:          bc.DB,
	}
}

// MineBlock creates a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) (*Block, error) {
	var lastHash []byte

	for _, tx := range transactions {
		success, err := bc.VerifyTransaction(tx)
		if err != nil {
			return nil, err
		}

		if !success {
			return nil, errors.New("invalid transaction")
		}
	}

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return nil, err
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		nbSer, err := newBlock.Serialize()
		if err != nil {
			return err
		}

		if err = b.Put(newBlock.Hash, nbSer); err != nil {
			return err
		}

		if err = b.Put([]byte("l"), newBlock.Hash); err != nil {
			return err
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newBlock, nil
}

// SignTransaction signs the inputs of a transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return err
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)

	return nil
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return false, err
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
