package blockchain

import (
	"github.com/boltdb/bolt"
)

// BIterator provides an iterator that allows us to retrieve each
// block in the blockchain
type BIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// Next retrieves the next block from the iterator
func (i *BIterator) Next() (*Block, error) {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		encodedBlock := b.Get(i.currentHash)
		bl, err := DeserializeBlock(encodedBlock)
		if err != nil {
			return err
		}

		block = bl

		return nil
	})
	if err != nil {
		return nil, err
	}

	i.currentHash = block.PrevBlockHash

	return block, nil
}
