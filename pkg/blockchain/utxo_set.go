package blockchain

import (
	"encoding/hex"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspendOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs, err := DeserializeOutputs(v)
			if err != nil {
				return err
			}

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspendOutputs[txID] = append(unspendOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})

	return accumulated, unspendOutputs, err
}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) ([]*TXOutput, error) {
	var UTXOs []*TXOutput
	db := u.Blockchain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs, err := DeserializeOutputs(v)
			if err != nil {
				return nil
			}

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})

	return UTXOs, err
}

// Reindex rebuilds the UTXO set
func (u UTXOSet) Reindex() error {
	db := u.Blockchain.DB
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(bucketName); err != nil && err != bolt.ErrBucketNotFound {
			return err
		}

		_, err := tx.CreateBucket(bucketName)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	UTXO, err := u.Blockchain.FindUTXO()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}

			ser, err := outs.Serialize()
			if err != nil {
				return err
			}

			if err = b.Put(key, ser); err != nil {
				return err
			}
		}
		return nil
	})
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet) Update(block *Block) error {
	db := u.Blockchain.DB

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs, err := DeserializeOutputs(outsBytes)
					if err != nil {
						return err
					}

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							return err
						}
					} else {
						ser, err := updatedOuts.Serialize()
						if err != nil {
							return nil
						}

						err = b.Put(vin.Txid, ser)
						if err != nil {
							return err
						}
					}
				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			ser, err := newOutputs.Serialize()
			if err != nil {
				return err
			}

			err = b.Put(tx.ID, ser)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
