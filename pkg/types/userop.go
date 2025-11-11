package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

var PACKED_USEROP_TYPEHASH common.Hash = crypto.Keccak256Hash(
	[]byte("PackedUserOperation(address sender,uint256 nonce,bytes initCode,bytes callData,bytes32 accountGasLimits,uint256 preVerificationGas,bytes32 gasFees,bytes paymasterAndData)"),
)

func (userOp *UserOperation) Hash(entryPoint common.Address, chainID *big.Int) common.Hash {
	// Pack user operation
	packed := userOp.Pack()

	// Create the hash: keccak256(abi.encode(packedUserOp, entryPoint, chainId))
	packedUserOpHash := crypto.Keccak256Hash(
		PACKED_USEROP_TYPEHASH[:],
		common.LeftPadBytes(packed.Sender[:], 32),
		common.LeftPadBytes(packed.Nonce.Bytes(), 32),
		crypto.Keccak256Hash(packed.InitCode).Bytes(),
		crypto.Keccak256Hash(packed.CallData).Bytes(),
		packed.AccountGasLimits[:],
		common.LeftPadBytes(packed.PreVerificationGas.Bytes(), 32),
		packed.GasFees[:],
		crypto.Keccak256Hash(packed.PaymasterAndData).Bytes(),
	)

	finalHash := crypto.Keccak256Hash(
		common.LeftPadBytes(packedUserOpHash[:], 32),
		common.LeftPadBytes(entryPoint[:], 32),
		common.LeftPadBytes(chainID.Bytes(), 32),
	)

	return finalHash
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

func (userOp *UserOperation) UnmarshalJSON(data []byte) error {
	// Intermediate struct with string fields
	type IntermediateUserOperation struct {
		Sender                        string `json:"sender"`
		Nonce                         string `json:"nonce"`
		Factory                       string `json:"factory"`
		FactoryData                   string `json:"factoryData"`
		CallData                      string `json:"callData"`
		CallGasLimit                  string `json:"callGasLimit"`
		VerificationGasLimit          string `json:"verificationGasLimit"`
		PreVerificationGas            string `json:"preVerificationGas"`
		MaxFeePerGas                  string `json:"maxFeePerGas"`
		MaxPriorityFeePerGas          string `json:"maxPriorityFeePerGas"`
		Paymaster                     string `json:"paymaster"`
		PaymasterVerificationGasLimit string `json:"paymasterVerificationGasLimit"`
		PaymasterPostOpGasLimit       string `json:"paymasterPostOpGasLimit"`
		PaymasterData                 string `json:"paymasterData"`
		Signature                     string `json:"signature"`
	}

	var imd IntermediateUserOperation
	if err := json.Unmarshal(data, &imd); err != nil {
		return err
	}

	// Parse addresses
	userOp.Sender = common.HexToAddress(imd.Sender)
	userOp.Factory = common.HexToAddress(imd.Factory)
	userOp.Paymaster = common.HexToAddress(imd.Paymaster)

	// Parse BigInts
	var ok bool
	userOp.Nonce, ok = new(big.Int).SetString(strings.TrimPrefix(imd.Nonce, "0x"), 16)
	if !ok {
		return fmt.Errorf("error unmarshalling nonce")
	}
	userOp.CallGasLimit, ok = new(big.Int).SetString(strings.TrimPrefix(imd.CallGasLimit, "0x"), 16)
	if !ok {
		return fmt.Errorf("error unmarshalling callGasLimit")
	}
	userOp.VerificationGasLimit, ok = new(big.Int).SetString(strings.TrimPrefix(imd.VerificationGasLimit, "0x"), 16)
	if !ok {
		return fmt.Errorf("error unmarshalling verificationGasLimit")
	}
	userOp.PreVerificationGas, ok = new(big.Int).SetString(strings.TrimPrefix(imd.PreVerificationGas, "0x"), 16)
	if !ok {
		return fmt.Errorf("error unmarshalling preVerificationGas")
	}
	userOp.MaxFeePerGas, ok = new(big.Int).SetString(strings.TrimPrefix(imd.MaxFeePerGas, "0x"), 16)
	if !ok {
		return fmt.Errorf("error unmarshalling maxFeePerGas")
	}
	userOp.MaxPriorityFeePerGas, ok = new(big.Int).SetString(strings.TrimPrefix(imd.MaxPriorityFeePerGas, "0x"), 16)
	if !ok {
		return fmt.Errorf("error unmarshalling maxPriorityFeePerGas")
	}

	// Parse byte arrays
	var err error
	userOp.FactoryData, err = hexutil.Decode(imd.FactoryData)
	if err != nil {
		return fmt.Errorf("error unmarshalling factoryData: %w", err)
	}
	userOp.CallData, err = hexutil.Decode(imd.CallData)
	if err != nil {
		return fmt.Errorf("error unmarshalling callData: %w", err)
	}
	userOp.PaymasterData, err = hexutil.Decode(imd.PaymasterData)
	if err != nil {
		return fmt.Errorf("error unmarshalling paymasterData: %w", err)
	}
	userOp.Signature, err = hexutil.Decode(imd.Signature)
	if err != nil {
		return fmt.Errorf("error unmarshalling signature: %w", err)
	}

	return nil
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
