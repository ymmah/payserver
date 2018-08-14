package crpc

import (
	"github.com/bitlum/connector/estimator"
	"github.com/bitlum/connector/metrics/rpc"
	core "github.com/bitlum/viabtc_rpc_client"
	"golang.org/x/net/context"
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/common"
	"github.com/go-errors/errors"
)

// defaultAccount default account which will be used for all request until
// account would be working properly for all assets.
var defaultAccount = "zigzag"

const (
	CreateReceipt       = "CreateReceipt"
	AccountAddress      = "AccountAddress"
	PendingBalance      = "PendingBalance"
	PendingTransactions = "PendingTransactions"
	GenerateTransaction = "GenerateTransaction"
	CheckReachable      = "CheckReachable"
	SendTransaction     = "SendTransaction"
	NetworkInfo         = "NetworkInfo"
	CreateInvoice       = "CreateInvoice"
	SendPayment         = "SendPayment"
	AccountBalance      = "AccountBalance"
)

// Server...
type Server struct {
	net                  string
	blockchainConnectors map[core.AssetType]connectors.BlockchainConnector
	lightningConnectors  map[core.AssetType]connectors.LightningConnector
	paymentsStore        common.PaymentsStore
	estmtr               estimator.USDEstimator
	metrics              rpc.MetricsBackend
}

// A compile time check to ensure that Server fully implements the
// ExchangeServer gRPC service.
var _ PayServerServer = (*Server)(nil)

// NewRPCServer creates and returns a new instance of the Server.
func NewRPCServer(net string,
	blockchainConnectors map[core.AssetType]connectors.BlockchainConnector,
	lightningConnectors map[core.AssetType]connectors.LightningConnector,
	paymentsStore common.PaymentsStore,
	estmtr estimator.USDEstimator,
	metrics rpc.MetricsBackend) (*Server, error) {
	return &Server{
		blockchainConnectors: blockchainConnectors,
		lightningConnectors:  lightningConnectors,
		paymentsStore:        paymentsStore,
		estmtr:               estmtr,
		metrics:              metrics,
		net:                  net,
	}, nil
}

//
// CreateReceipt is used to create blockchain deposit address in
// case of blockchain media, and lightning network invoice in
// case of the lightning media, which will be used to receive money from
// external entity.
func (s *Server) CreateReceipt(ctx context.Context,
	req *CreateReceiptRequest) (*CreateReceiptResponse, error) {

	log.Tracef("command(%v), request(%v)", getFunctionName(),
		convertProtoMessage(req))

	var resp *CreateReceiptResponse

	switch req.Media {
	case Media_BLOCKCHAIN:
		c, ok := s.blockchainConnectors[core.AssetType(req.Asset.String())]
		if !ok {
			severity := errMetricsInfo(ErrAssetNotSupported)
			s.metrics.AddError(CreateReceipt, severity)
			return nil, newErrAssetNotSupported(req.Asset.String())
		}

		address, err := c.CreateAddress(defaultAccount)
		if err != nil {
			return nil, newErrInternal(err.Error())
		}

		resp = &CreateReceiptResponse{
			Receipt: address,
		}
	case Media_LIGHTNING:
		c, ok := s.lightningConnectors[core.AssetType(req.Asset.String())]
		if !ok {
			severity := errMetricsInfo(ErrNetworkNotSupported)
			s.metrics.AddError(CreateReceipt, severity)
			return nil, newErrAssetNotSupported(req.Asset.String())
		}

		// Ensure that even if amount is not specified we treat it as zero
		// value.
		if req.Amount == "" {
			req.Amount = "0"
		}

		invoice, err := c.CreateInvoice(defaultAccount, req.Amount)
		if err != nil {
			return nil, newErrInternal(err.Error())
		}

		resp = &CreateReceiptResponse{
			Receipt: invoice,
		}
	default:
		return nil, errors.Errorf("media(%v) is not supported",
			req.Media.String())
	}

	log.Tracef("command(%v), response(%v)", getFunctionName(),
		convertProtoMessage(resp))

	return resp, nil
}

//
// ValidateReceipt is used to validate receipt for given asset and media.
func (s *Server) ValidateReceipt(ctx context.Context,
	req *ValidateReceiptRequest) (*EmptyResponse, error) {
	log.Tracef("command(%v), request(%v)", getFunctionName(),
		convertProtoMessage(req))

	switch req.Media {
	case Media_BLOCKCHAIN:
		c, ok := s.blockchainConnectors[core.AssetType(req.Asset.String())]
		if !ok {
			severity := errMetricsInfo(ErrAssetNotSupported)
			s.metrics.AddError(CreateReceipt, severity)
			return nil, newErrAssetNotSupported(req.Asset.String())
		}

		if err := c.ValidateAddress(req.Receipt); err != nil {
			return nil, err
		}

	case Media_LIGHTNING:
		c, ok := s.lightningConnectors[core.AssetType(req.Asset.String())]
		if !ok {
			severity := errMetricsInfo(ErrNetworkNotSupported)
			s.metrics.AddError(CreateReceipt, severity)
			return nil, newErrAssetNotSupported(req.Asset.String())
		}

		if err := c.ValidateInvoice(req.Receipt, req.Amount); err != nil {
			return nil, err
		}

	default:
		return nil, errors.Errorf("media(%v) is not supported",
			req.Media.String())
	}

	resp := &EmptyResponse{}
	log.Tracef("command(%v), response(%v)", getFunctionName(),
		convertProtoMessage(resp))

	return resp, nil
}

//
// Balance is used to determine balance.
func (s *Server) Balance(ctx context.Context, in *BalanceRequest,
) (*BalanceResponse, error) {
	return &BalanceResponse{}, nil
}

//
// EstimateFee estimates the fee of the payment.
func (s *Server) EstimateFee(ctx context.Context, in *EstimateFeeRequest) (*EstimateFeeResponse, error) {
	return &EstimateFeeResponse{}, nil
}

//
// SendPayment sends payment to the given recipient,
// ensures in the validity of the receipt as well as the
// account has enough money for doing that.
func (s *Server) SendPayment(ctx context.Context, in *SendPaymentRequest) (*Payment, error) {
	return &Payment{}, nil
}

//
// PaymentByID is used to fetch the information about payment, by the
// given system payment id.
func (s *Server) PaymentByID(ctx context.Context, in *PaymentByIDRequest) (*Payment, error) {
	return &Payment{}, nil
}

//
// PaymentsByReceipt is used to fetch the information about payment, by the
// given receipt.
func (s *Server) PaymentsByReceipt(ctx context.Context,
	in *PaymentsByReceiptRequest) (*PaymentsByReceiptResponse, error) {
	return &PaymentsByReceiptResponse{}, nil
}

//
// ListPayments returns list of payment which were registered by the
// system.
func (s *Server) ListPayments(ctx context.Context,
	in *ListPaymentsRequest) (*ListPaymentsResponse, error) {
	return &ListPaymentsResponse{}, nil
}

////
//// AccountAddress return the deposit address of account.
////
//// NOTE: Works only for blockchain daemons.
//func (s *Server) AccountAddress(_ context.Context,
//	req *AccountAddressRequest) (*Address, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
//	if !ok {
//		severity := errMetricsInfo(ErrAssetNotSupported)
//		s.metrics.AddError(AccountAddress, severity)
//		return nil, newErrAssetNotSupported(req.Asset, "account address")
//	}
//
//	address, err := c.AccountAddress(req.Account)
//	if err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	resp := &Address{
//		Data: address,
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// PendingBalance return the amount of funds waiting to be confirmed.
////
//// NOTE: Works only for blockchain daemons.
//func (s *Server) PendingBalance(_ context.Context,
//	req *PendingBalanceRequest) (*Balance, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
//	if !ok {
//		severity := errMetricsInfo(ErrAssetNotSupported)
//		s.metrics.AddError(PendingBalance, severity)
//		return nil, newErrAssetNotSupported(req.Asset, "pending balance")
//	}
//
//	balance, err := c.PendingBalance(req.Account)
//	if err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	resp := &Balance{
//		Data: balance,
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// PendingTransactions return the transactions which has confirmation
//// number lower the required by payment system.
////
//// NOTE: Works only for blockchain daemons.
//func (s *Server) PendingTransactions(_ context.Context,
//	req *PendingTransactionsRequest) (*PendingTransactionsResponse, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
//	if !ok {
//		severity := errMetricsInfo(ErrAssetNotSupported)
//		s.metrics.AddError(PendingTransactions, severity)
//		return nil, newErrAssetNotSupported(req.Asset, "pending transactions")
//	}
//
//	txs, err := c.PendingTransactions(req.Account)
//	if err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	payments := make([]*BlockchainPendingPayment, len(txs))
//	for i, tx := range txs {
//		payments[i] = &BlockchainPendingPayment{
//			Payment: &Payment{
//				Id:      tx.ID,
//				Amount:  tx.Amount.String(),
//				Account: tx.Account,
//				Address: tx.Address,
//				Type:    string(tx.Type),
//			},
//			Confirmations:     tx.Confirmations,
//			ConfirmationsLeft: tx.ConfirmationsLeft,
//		}
//	}
//
//	resp := &PendingTransactionsResponse{
//		Payments: payments,
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// GenerateTransaction generates raw blockchain transaction.
////
//// NOTE: Blockchain endpoint.
//func (s *Server) GenerateTransaction(_ context.Context,
//	req *GenerateTransactionRequest) (*GenerateTransactionResponse, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
//	if !ok {
//		severity := errMetricsInfo(ErrAssetNotSupported)
//		s.metrics.AddError(GenerateTransaction, severity)
//		return nil, newErrAssetNotSupported(req.Asset, "generate transaction")
//	}
//
//	genTx, err := c.GenerateTransaction(req.ReceiverAddress, req.Amount)
//	if err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	resp := &GenerateTransactionResponse{
//		RawTx: genTx.Bytes(),
//		TxId:  genTx.ID(),
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// SendTransaction send the given transaction to the blockchain network.
////
//// NOTE: Works only for blockchain daemons.
//func (s *Server) SendTransaction(_ context.Context,
//	req *SendTransactionRequest) (*EmtpyResponse, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
//	if !ok {
//		severity := errMetricsInfo(ErrAssetNotSupported)
//		s.metrics.AddError(SendTransaction, severity)
//		return nil, newErrAssetNotSupported(req.Asset, "send transaction")
//	}
//
//	if err := c.SendTransaction(req.RawTx); err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	resp := &EmtpyResponse{}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// NetworkInfo returns information about the daemon and its network,
//// depending on the requested
//func (s *Server) Info(_ context.Context,
//	req *InfoRequest) (*InfoResponse, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	var net Net
//	switch s.net {
//	case "simnet":
//		net = Net_simnet
//	case "testnet":
//		net = Net_testnet
//	case "mainnet":
//		net = Net_mainnet
//	}
//
//	var linfo *LightningInfo
//	c, ok := s.lightningConnectors[core.AssetType("BTC")]
//	if ok {
//		info, err := c.Info()
//		if err != nil {
//			return nil, newErrInternal(err.Error())
//		}
//
//		linfo = &LightningInfo{
//			Host:               info.Host,
//			Port:               info.Port,
//			MinAmount:          info.MinAmount,
//			MaxAmount:          info.MaxAmount,
//			IdentityPubkey:     info.IdentityPubkey,
//			Alias:              info.Alias,
//			NumPendingChannels: info.NumPendingChannels,
//			NumActiveChannels:  info.NumActiveChannels,
//			NumPeers:           info.NumPeers,
//			BlockHeight:        info.BlockHeight,
//			BlockHash:          info.BlockHash,
//		}
//	}
//
//	resp := &InfoResponse{
//		Time:         time.Now().String(),
//		Net:          net,
//		LightingInfo: linfo,
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// CreateInvoice creates recept for sender lightning node which contains
//// the information about receiver node and
////
//// NOTE: Works only for lightning network daemons.
//func (s *Server) CreateInvoice(_ context.Context,
//	req *CreateInvoiceRequest) (*Invoice, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	c, ok := s.lightningConnectors[core.AssetType(req.Asset)]
//	if !ok {
//		severity := errMetricsInfo(ErrNetworkNotSupported)
//		s.metrics.AddError(CreateInvoice, severity)
//		return nil, newErrAssetNotSupported(req.Asset, "create invoice")
//	}
//
//	invoice, err := c.CreateInvoice(req.Account, req.Amount)
//	if err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	resp := &Invoice{
//		Data: invoice,
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// SendPayment is used to send specific amount of money inside lightning
//// network.
////
//// NOTE: Works only for lightning network daemons.
//func (s *Server) SendPayment(_ context.Context,
//	req *SendPaymentRequest) (*EmtpyResponse, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	c, ok := s.lightningConnectors[core.AssetType(req.Asset)]
//	if !ok {
//		severity := errMetricsInfo(ErrAssetNotSupported)
//		s.metrics.AddError(SendPayment, severity)
//		return nil, newErrAssetNotSupported(req.Asset, "send payment")
//	}
//
//	if err := c.SendTo(req.Invoice); err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	resp := &EmtpyResponse{}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// CheckReachable checks that given node can be reached from our
//// lightning node.
////
//// NOTE: Works only for lightning network daemons.
//func (s *Server) CheckReachable(_ context.Context,
//	req *CheckReachableRequest) (*CheckReachableResponse, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	c, ok := s.lightningConnectors[core.AssetType(req.Asset)]
//	if !ok {
//		severity := errMetricsInfo(ErrAssetNotSupported)
//		s.metrics.AddError(CheckReachable, severity)
//		return nil, newErrAssetNotSupported(req.Asset, "create invoice")
//	}
//
//	amount := "0.00000001"
//	routes, err := c.QueryRoutes(req.IdentityKey, amount, 1)
//	if err != nil {
//		// TODO(andrew.shvv) distinguish errors
//		return &CheckReachableResponse{
//			IsReachable: false,
//		}, nil
//	}
//
//	resp := &CheckReachableResponse{}
//	if len(routes) != 0 {
//		resp.IsReachable = true
//	} else {
//		resp.IsReachable = false
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// Estimate estimates the dollar price of the choosen asset.
//func (s *Server) Estimate(_ context.Context,
//	req *EstimateRequest) (*EstimationResponse, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	usdEstimation, err := s.estmtr.Estimate(req.Asset, req.Amount)
//	if err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	resp := &EstimationResponse{
//		Usd: usdEstimation,
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
//// PaymentReceived is used to determine if payment with given ID is received
//func (s *Server) PaymentReceived(_ context.Context,
//	req *PaymentReceivedRequest) (*PaymentReceivedResponse, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))
//
//	_, err := s.paymentsStore.Payment(req.Id)
//
//	if err != nil && err != common.PaymentNotFound {
//		return nil, err
//	}
//
//	resp := &PaymentReceivedResponse{
//		Received: err == nil,
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}
//
////
//// AccountBalance is used to determine account balance state for
//// requested asset. This state includes available and pending balance.
////
//// NOTE: Works only for blockchain daemons.
//func (s *Server) AccountBalance(_ context.Context,
//	req *AccountBalanceRequest) (*AccountBalanceResponse, error) {
//
//	log.Tracef("command(%v), request(%v)", getFunctionName(),
//		spew.Sdump(req))
//
//	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
//	if !ok {
//		severity := errMetricsInfo(ErrAssetNotSupported)
//		s.metrics.AddError(AccountBalance, severity)
//		return nil, newErrAssetNotSupported(
//			req.Asset, "account balance")
//	}
//
//	available, err := c.ConfirmedBalance(req.Account)
//	if err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	pending, err := c.PendingBalance(req.Account)
//	if err != nil {
//		return nil, newErrInternal(err.Error())
//	}
//
//	resp := &AccountBalanceResponse{
//		Available: available,
//		Pending:   pending,
//	}
//
//	log.Tracef("command(%v), response(%v)", getFunctionName(),
//		spew.Sdump(resp))
//
//	return resp, nil
//}