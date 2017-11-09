package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	perrors "github.com/pkg/errors"
)

const subsidy = 10

// Transaction represents a blockchain transaction
type Transaction struct {
	ID   []byte
	Vin  []*TXInput
	Vout []*TXOutput
}

// IsCoinbase checks whether the transaction is a coinbase transaction
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Serialize serializes the transaction
func (tx Transaction) Serialize() ([]byte, error) {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		return nil, perrors.Wrap(err, "failed to encode transaction")
	}

	return encoded.Bytes(), nil
}

// Hash returns the hash of the transaction
func (tx *Transaction) Hash() ([]byte, error) {
	txCopy := *tx
	txCopy.ID = []byte{}

	ser, err := tx.Serialize()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to serialize transaction")
	}

	hash := sha256.Sum256(ser)

	return hash[:], nil
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

// NewCoinbaseTransaction creates a new coinbase transaction
func NewCoinbaseTransaction(to, data string) (*Transaction, error) {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := &TXInput{
		Txid:      []byte{},
		Vout:      -1,
		Signature: nil,
		PubKey:    []byte(data),
	}

	txout := NewTXOutput(subsidy, to)

	tx := Transaction{
		ID:   nil,
		Vin:  []*TXInput{txin},
		Vout: []*TXOutput{txout},
	}

	id, err := tx.Hash()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to hash transaction")
	}
	tx.ID = id

	return &tx, nil
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) (*Transaction, error) {
	var inputs []*TXInput
	var outputs []*TXOutput

	wallets, err := NewWallets()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to create new wallets")
	}

	wallet := wallets.GetWallet(from)
	pubKeyHash, err := HashPublicKey(wallet.PublicKey)
	if err != nil {
		return nil, perrors.Wrap(err, "failed to hash public key of wallet")
	}

	acc, validOutputs, err := bc.FindSpendableOutputs(pubKeyHash, amount)
	if err != nil {
		return nil, perrors.Wrap(err, "failed to find spendable outputs")
	}

	if acc < amount {
		return nil, errors.New("not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			return nil, perrors.Wrap(err, "failed to decode transaction id")
		}

		for _, out := range outs {
			input := &TXInput{
				Txid:      txID,
				Vout:      out,
				Signature: nil,
				PubKey:    wallet.PublicKey,
			}

			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, NewTXOutput(acc-amount, from)) // change
	}

	tx := &Transaction{
		ID:   nil,
		Vin:  inputs,
		Vout: outputs,
	}

	id, err := tx.Hash()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to hash transaction")
	}
	tx.ID = id

	return tx, nil
}
