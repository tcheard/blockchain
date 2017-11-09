package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
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
		return nil, err
	}

	return encoded.Bytes(), nil
}

// Hash returns the hash of the transaction
func (tx *Transaction) Hash() ([]byte, error) {
	txCopy := *tx
	txCopy.ID = []byte{}

	ser, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(ser)

	return hash[:], nil
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			return errors.New("previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		h, err := txCopy.Hash()
		if err != nil {
			return err
		}
		txCopy.ID = h
		txCopy.Vin[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			return err
		}

		sig := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = sig
	}

	return nil
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

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []*TXInput
	var outputs []*TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, &TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, &TXOutput{vout.Value, vout.PubKeyHash})
	}

	return Transaction{
		ID:   tx.ID,
		Vin:  inputs,
		Vout: outputs,
	}
}

// Verify verifies the signatures of the transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			return false, errors.New("previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		h, err := txCopy.Hash()
		if err != nil {
			return false, err
		}
		txCopy.ID = h
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{
			Curve: curve,
			X:     &x,
			Y:     &y,
		}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false, nil
		}
	}

	return true, nil
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
		return nil, err
	}
	tx.ID = id

	return &tx, nil
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(from, to string, amount int, UTXOSet *UTXOSet) (*Transaction, error) {
	var inputs []*TXInput
	var outputs []*TXOutput

	wallets, err := NewWallets()
	if err != nil {
		return nil, err
	}

	wallet := wallets.GetWallet(from)
	pubKeyHash, err := HashPublicKey(wallet.PublicKey)
	if err != nil {
		return nil, err
	}

	acc, validOutputs, err := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)
	if err != nil {
		return nil, err
	}

	if acc < amount {
		return nil, errors.New("not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			return nil, err
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
		return nil, err
	}
	tx.ID = id

	err = UTXOSet.Blockchain.SignTransaction(tx, wallet.PrivateKey)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
