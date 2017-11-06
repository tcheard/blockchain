package main

import (
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

// BlockchainIterator provides an iterator that allows us to retrieve each
// block in the blockchain
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// Iterator retrieves an iterator for the blockchain
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		currentHash: bc.tip,
		db:          bc.db,
	}
}

// Next retrieves the next block from the iterator
func (i *BlockchainIterator) Next() (*Block, error) {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		encodedBlock := b.Get(i.currentHash)
		bl, err := DeserializeBlock(encodedBlock)
		if err != nil {
			return perrors.Wrap(err, "failed to deserialize block that was read")
		}

		block = bl

		return nil
	})
	if err != nil {
		return nil, perrors.Wrap(err, "failed to retrieve next block")
	}

	i.currentHash = block.PrevBlockHash

	return block, nil
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain creates a new blockchain by reading from the database
func NewBlockchain() (*Blockchain, error) {
	if dbExists() == false {
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
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
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
