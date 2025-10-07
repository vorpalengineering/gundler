package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type UserOperation struct {
	Sender                        common.Address `json:"sender"`
	Nonce                         *big.Int       `json:"nonce"`
	Factory                       common.Address `json:"factory"`
	FactoryData                   []byte         `json:"factoryData"`
	CallData                      []byte         `json:"callData"`
	CallGasLimit                  *big.Int       `json:"callGasLimit"`
	VerificationGasLimit          *big.Int       `json:"verificationGasLimit"`
	PreVerificationGas            *big.Int       `json:"preVerificationGas"`
	MaxFeePerGas                  *big.Int       `json:"maxFeePerGas"`
	MaxPriorityFeePerGas          *big.Int       `json:"maxPriorityFeePerGas"`
	Paymaster                     common.Address `json:"paymaster"`
	PaymasterVerificationGasLimit *big.Int       `json:"paymasterVerificationGasLimit"`
	PaymasterPostOpGasLimit       *big.Int       `json:"paymasterPostOpGasLimit"`
	PaymasterData                 []byte         `json:"paymasterData"`
	Signature                     []byte         `json:"signature"`
}

type PackedUserOperation struct {
	Sender             common.Address `json:"sender"`
	Nonce              *big.Int       `json:"nonce"`
	InitCode           []byte         `json:"initCode"`
	CallData           []byte         `json:"callData"`
	AccountGasLimits   [32]byte       `json:"accountGasLimits"`
	PreVerificationGas *big.Int       `json:"preVerificationGas"`
	GasFees            [32]byte       `json:"gasFees"`
	PaymasterAndData   []byte         `json:"paymasterAndData"`
	Signature          []byte         `json:"signature"`
}

// TODO: implement
func (userOp *UserOperation) Hash() common.Hash {
	return common.Hash{}
}

func (userOp *UserOperation) Pack() *PackedUserOperation {
	return &PackedUserOperation{
		Sender:             userOp.Sender,
		Nonce:              userOp.Nonce,
		InitCode:           append(userOp.Factory[:], userOp.FactoryData...),
		CallData:           userOp.CallData,
		AccountGasLimits:   packAccountGasLimits(userOp.VerificationGasLimit, userOp.CallGasLimit),
		PreVerificationGas: userOp.PreVerificationGas,
		GasFees:            packGasFees(userOp.MaxPriorityFeePerGas, userOp.MaxFeePerGas),
		PaymasterAndData: packPaymasterAndData(
			userOp.Paymaster,
			userOp.PaymasterVerificationGasLimit,
			userOp.PaymasterPostOpGasLimit,
			userOp.PaymasterData,
		),
		Signature: userOp.Signature,
	}
}

func packAccountGasLimits(verificationGasLimit *big.Int, callGasLimit *big.Int) [32]byte {
	var accountGasLimits [32]byte

	// Convert to bytes for concatenation
	verificationGasLimitBytes := verificationGasLimit.Bytes()
	callGasLimitBytes := callGasLimit.Bytes()

	// Copy into byte array (right-aligned)
	copy(accountGasLimits[16-len(verificationGasLimitBytes):16], verificationGasLimitBytes)
	copy(accountGasLimits[32-len(callGasLimitBytes):32], callGasLimitBytes)

	return accountGasLimits
}

func packGasFees(maxPriorityFeePerGas *big.Int, maxFeePerGas *big.Int) [32]byte {
	var gasFees [32]byte

	// Convert to bytes for concatenation
	maxPriorityFeePerGasBytes := maxPriorityFeePerGas.Bytes()
	maxFeePerGasBytes := maxFeePerGas.Bytes()

	// Copy into byte array (right-aligned)
	copy(gasFees[16-len(maxPriorityFeePerGasBytes):16], maxPriorityFeePerGasBytes)
	copy(gasFees[32-len(maxFeePerGasBytes):32], maxFeePerGasBytes)

	return gasFees
}

func packPaymasterAndData(
	paymaster common.Address,
	paymasterVerificationGasLimit *big.Int,
	paymasterPostOpGasLimit *big.Int,
	paymasterData []byte,
) []byte {
	// Return empty byte array if no paymaster address
	if (paymaster == common.Address{}) {
		return []byte{}
	}

	// Bytes 0-19: Paymaster address (20 bytes)
	// Bytes 20-35: PaymasterVerificationGasLimit (16 bytes)
	// Bytes 36-51: PaymasterPostOpGasLimit (16 bytes)
	// Bytes 52+: PaymasterData (variable)
	var paymasterAndData []byte

	// Convert to bytes for concatenation
	paymasterBytes := paymaster.Bytes()
	paymasterVerificationGasLimitBytes := paymasterVerificationGasLimit.Bytes()
	paymasterPostOpGasLimitBytes := paymasterPostOpGasLimit.Bytes()

	// Copy into byte array (right-aligned)
	copy(paymasterAndData[20-len(paymasterBytes):20], paymasterBytes)
	copy(paymasterAndData[35-len(paymasterVerificationGasLimitBytes):35], paymasterVerificationGasLimitBytes)
	copy(paymasterAndData[51-len(paymasterPostOpGasLimitBytes):51], paymasterPostOpGasLimitBytes)
	paymasterAndData = append(paymasterAndData, paymasterData...)

	return paymasterAndData
}
