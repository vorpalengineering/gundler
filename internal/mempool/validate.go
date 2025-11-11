package mempool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vorpalengineering/gundler/pkg/types"
)

func (pool *Mempool) validateUserOp(userOp *types.UserOperation) error {
	// Check sender
	if userOp.Sender == (common.Address{}) {
		return fmt.Errorf("invalid sender")
	}

	// Check nonce
	if userOp.Nonce == nil || userOp.Nonce.Sign() < 0 {
		return fmt.Errorf("invalid nonce")
	}

	// Check gas limits
	if userOp.CallGasLimit == nil || userOp.CallGasLimit.Sign() <= 0 {
		return fmt.Errorf("invalid callGasLimit")
	}
	if userOp.VerificationGasLimit == nil || userOp.VerificationGasLimit.Sign() <= 0 {
		return fmt.Errorf("invalid verificationGasLimit")
	}
	if userOp.PreVerificationGas == nil || userOp.PreVerificationGas.Sign() <= 0 {
		return fmt.Errorf("invalid preVerificationGas")
	}

	// Check fees
	if userOp.MaxFeePerGas == nil || userOp.MaxFeePerGas.Sign() <= 0 {
		return fmt.Errorf("invalid maxFeePerGas")
	}
	if userOp.MaxPriorityFeePerGas == nil || userOp.MaxPriorityFeePerGas.Sign() < 0 {
		return fmt.Errorf("invalid maxPriorityFeePerGas")
	}
	if userOp.MaxPriorityFeePerGas.Cmp(userOp.MaxFeePerGas) > 0 {
		return fmt.Errorf("maxPriorityFeePerGas cannot exceed maxFeePerGas")
	}

	// Check signature
	if len(userOp.Signature) == 0 {
		return fmt.Errorf("signature is required")
	}

	// Check paymaster if set
	if userOp.Paymaster != (common.Address{}) {
		if userOp.PaymasterVerificationGasLimit == nil || userOp.PaymasterVerificationGasLimit.Sign() <= 0 {
			return fmt.Errorf("invalid paymasterVerificationGasLimit")
		}
		if userOp.PaymasterPostOpGasLimit == nil || userOp.PaymasterPostOpGasLimit.Sign() < 0 {
			return fmt.Errorf("invalid paymasterPostOpGasLimit")
		}
	}

	return nil
}
