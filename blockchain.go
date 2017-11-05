package main

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

const (
	dbFile       = "blockchain.db"
	blocksBucket = "blocks"
)

// Blockchain represents the actual blockchain holding all its blocks
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// NewBlockchain creates a new blockchain
func NewBlockchain() (*Blockchain, error) {
	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open database file")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			fmt.Println("No existing blockchain found. Creating a new one...")

			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				return errors.Wrap(err, "Failed to create database bucket")
			}

			genSer, err := genesis.Serialize()
			if err != nil {
				return errors.Wrap(err, "Failed to serialize genesis block")
			}

			err = b.Put(genesis.Hash, genSer)
			if err != nil {
				return errors.Wrap(err, "Failed to put the genesis block")
			}

			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				return errors.Wrap(err, "Failed to put the 'l' key")
			}

			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "Failed to update bucket")
	}

	return &Blockchain{tip: tip, db: db}, nil
}

// AddBlock adds a block to the Blockchain
func (bc *Blockchain) AddBlock(data string) error {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve last hash")
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		nbSer, err := newBlock.Serialize()
		if err != nil {
			return errors.Wrap(err, "Failed to serialize new block")
		}

		err = b.Put(newBlock.Hash, nbSer)
		if err != nil {
			return errors.Wrap(err, "Failed to put the new block")
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return errors.Wrap(err, "Failed to put the 'l' key")
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "Failed to write new block")
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
			return errors.Wrap(err, "Failed to deserialize block that was read")
		}

		block = bl

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve next block")
	}

	i.currentHash = block.PrevBlockHash

	return block, nil
}
