package blockchain

import (
	"github.com/boltdb/bolt"
	perrors "github.com/pkg/errors"
)

// BlockchainIterator provides an iterator that allows us to retrieve each
// block in the blockchain
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
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
