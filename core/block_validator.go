// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
)

/** added this section for purposing**/

/** end this section **/

// client, err       = ethclient.Dial("/home/thach/.ethereum/geth.ipc")
// BlockValidator is responsible for validating block headers, uncles and
// processed state.
//
// BlockValidator implements Validator.
type BlockValidator struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for validating
}

// NewBlockValidator returns a new block validator which is safe for re-use
func NewBlockValidator(config *params.ChainConfig, blockchain *BlockChain, engine consensus.Engine) *BlockValidator {
	validator := &BlockValidator{
		config: config,
		engine: engine,
		bc:     blockchain,
	}
	return validator
}

var (
	opts      *bind.CallOpts
	Contract1 *bind.BoundContract2
)

// ValidateBody validates the given block's uncles and verifies the the block
// header's transaction and uncle roots. The headers are assumed to be already
// validated at this point.
func (v *BlockValidator) ValidateBody(block *types.Block) error {
	// Check whether the block's known, and if not, that it's linkable
	if v.bc.HasBlockAndState(block.Hash()) {
		return ErrKnownBlock
	}
	if !v.bc.HasBlockAndState(block.ParentHash()) {
		return consensus.ErrUnknownAncestor
	}
	// checkingMinerAddr := `0x4AD87106FDA30C4FAd7c6230D09bF69F25d023eD`

	// return fmt.Errorf("you should get permission to mine in our network: have %x", header.Coinbase)
	// Header validity is known at this point, check the uncles and transactions
	header := block.Header()
	// ipcEnpoint := node.ReturnIPCendpoint()
	client, err := ethclient.Dial(node.IpcEndpointValue)
	log.Info(fmt.Sprintf("IPC endpoint opened: %s", node.IpcEndpointValue))
	checkingMinerABI := `[{"constant":false,"inputs":[{"name":"_miner","type":"address"}],"name":"removeMiner","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_miner","type":"address"}],"name":"check","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"miner","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_miner","type":"address"}],"name":"addMiner","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"}]`
	checkingMinerAddr := `0x5D48d85Cbad801d76523DD1A890af2B5ee18D08b`
	// checkingMinerAddr := []byte{0x4A, 0xD8, 0x71, 0x06, 0xFD, 0xA3, 0x0C, 0x4F, 0xAD, 0x7C, 0x62, 0x30, 0xD0, 0x9B, 0xF6, 0x9F, 0x25, 0xD0, 0x23, 0xED}
	parsed, err := abi.JSON(strings.NewReader(checkingMinerABI))
	if err != nil {
		return err
	}
	opts := new(bind.CallOpts)
	ctx := ensureContext2(opts.Context)
	// Contract1 = bind.NewBoundContract(common.BytesToAddress(checkingMinerAddr), parsed, nil, nil)
	Contract1 = &bind.BoundContract2{
		Address: common.HexToAddress(checkingMinerAddr),
		Abi:     parsed,
	}

	// output, err := Contract1.CallContractChecking(nil, "check", header.Coinbase)
	input, err := Contract1.Abi.Pack("check", header.Coinbase)
	if err != nil {
		log.Info("something wrong when call to get contract state")
		return err
	}
	allocData := []byte{0x3B, 0x58, 0xE3, 0xED, 0x47, 0xDA, 0x42, 0x2C, 0xFE, 0xEF, 0xE5, 0xEB, 0x47, 0xCA, 0x44, 0xE4, 0x3E, 0x37, 0x57, 0xE6}
	msg := ethereum.CallMsg{From: common.BytesToAddress(allocData), To: &Contract1.Address, Data: input}
	// var output hexutil.Bytes
	output, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		log.Info("something wrong when call to get contract state")
		return err
	}
	resultString := string(output)
	result := string(output[31] + 48)
	if err != nil {
		log.Info("something wrong when call to get contract state")
		return err
	}
	if result == "0" {
		// log.Info("" + fmt.Printf("Have out put string : "+s))
		log.Info("Have out put string : %t\n", result)
		return fmt.Errorf("you should get permission to mine in our network: have %x \n %x \n %t \n"+resultString, header.Coinbase, output, result)
	}
	if err := v.engine.VerifyUncles(v.bc, block); err != nil {
		return err
	}
	if hash := types.CalcUncleHash(block.Uncles()); hash != header.UncleHash {
		return fmt.Errorf("uncle root hash mismatch: have %x, want %x", hash, header.UncleHash)
	}
	if hash := types.DeriveSha(block.Transactions()); hash != header.TxHash {
		return fmt.Errorf("transaction root hash mismatch: have %x, want %x", hash, header.TxHash)
	}
	return nil
}

// ValidateState validates the various changes that happen after a state
// transition, such as amount of used gas, the receipt roots and the state root
// itself. ValidateState returns a database batch if the validation was a success
// otherwise nil and an error is returned.
func (v *BlockValidator) ValidateState(block, parent *types.Block, statedb *state.StateDB, receipts types.Receipts, usedGas *big.Int) error {
	header := block.Header()
	if block.GasUsed().Cmp(usedGas) != 0 {
		return fmt.Errorf("invalid gas used (remote: %v local: %v)", block.GasUsed(), usedGas)
	}
	// Validate the received block's bloom with the one derived from the generated receipts.
	// For valid blocks this should always validate to true.
	rbloom := types.CreateBloom(receipts)
	if rbloom != header.Bloom {
		return fmt.Errorf("invalid bloom (remote: %x  local: %x)", header.Bloom, rbloom)
	}
	// Tre receipt Trie's root (R = (Tr [[H1, R1], ... [Hn, R1]]))
	receiptSha := types.DeriveSha(receipts)
	if receiptSha != header.ReceiptHash {
		return fmt.Errorf("invalid receipt root hash (remote: %x local: %x)", header.ReceiptHash, receiptSha)
	}
	// Validate the state root against the received state root and throw
	// an error if they don't match.
	if root := statedb.IntermediateRoot(v.config.IsEIP158(header.Number)); header.Root != root {
		return fmt.Errorf("invalid merkle root (remote: %x local: %x)", header.Root, root)
	}
	return nil
}

// CalcGasLimit computes the gas limit of the next block after parent.
// The result may be modified by the caller.
// This is miner strategy, not consensus protocol.
func CalcGasLimit(parent *types.Block) *big.Int {
	// contrib = (parentGasUsed * 3 / 2) / 1024
	contrib := new(big.Int).Mul(parent.GasUsed(), big.NewInt(3))
	contrib = contrib.Div(contrib, big.NewInt(2))
	contrib = contrib.Div(contrib, params.GasLimitBoundDivisor)

	// decay = parentGasLimit / 1024 -1
	decay := new(big.Int).Div(parent.GasLimit(), params.GasLimitBoundDivisor)
	decay.Sub(decay, big.NewInt(1))

	/*
		strategy: gasLimit of block-to-mine is set based on parent's
		gasUsed value.  if parentGasUsed > parentGasLimit * (2/3) then we
		increase it, otherwise lower it (or leave it unchanged if it's right
		at that usage) the amount increased/decreased depends on how far away
		from parentGasLimit * (2/3) parentGasUsed is.
	*/
	gl := new(big.Int).Sub(parent.GasLimit(), decay)
	gl = gl.Add(gl, contrib)
	gl.Set(math.BigMax(gl, params.MinGasLimit))

	// however, if we're now below the target (TargetGasLimit) we increase the
	// limit as much as we can (parentGasLimit / 1024 -1)
	if gl.Cmp(params.TargetGasLimit) < 0 {
		gl.Add(parent.GasLimit(), decay)
		gl.Set(math.BigMin(gl, params.TargetGasLimit))
	}
	return gl
}

func ensureContext2(ctx context.Context) context.Context {
	if ctx == nil {
		return context.TODO()
	}
	return ctx
}
