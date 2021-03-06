package daemon

import (
	"math/big"
	"time"

	"strings"

	"strconv"

	"github.com/Rennbon/boxwallet/bccoin"
	"github.com/Rennbon/boxwallet/bccore"
	"github.com/Rennbon/boxwallet/bctrans/client"
	"github.com/Rennbon/boxwallet/errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type BtcDaemon struct {
	*client.BtcClient
	TickSecond time.Duration
	ConfirmNum uint64
	UsdtId     int
	UnlockNum  uint64
}

const (
	opReturnStr = "OP_RETURN"
	nulldata    = "nulldata"
	pubkeyhash  = "pubkeyhash"
)

var (
	btcBG *BtcDaemon
)

func NewBtcDaemon(confirmNum, unlocknNum uint64, tickSecond time.Duration) *BtcDaemon {
	bctCli, err := client.GetBtcClientIntance()

	if err != nil {
		panic(err)
	}
	omniCli, err := client.GetUsdtClientInstance()
	if err != nil {
		panic(err)
	}
	bd := &BtcDaemon{
		TickSecond: tickSecond,
		ConfirmNum: confirmNum,
		BtcClient:  bctCli,
		UsdtId:     omniCli.PropertyId,
		UnlockNum:  unlocknNum,
	}
	return bd
}

func (d *BtcDaemon) GetLoopDuration() time.Duration {
	return d.TickSecond
}

func (d *BtcDaemon) GetBlockHeight() (*big.Int, uint64, error) {
	bc, err := d.C.GetBlockChainInfo()
	if err != nil {
		return nil, 0, err
	}
	return big.NewInt(int64(bc.Blocks)), uint64(bc.Headers), nil
}

func (d *BtcDaemon) GetBlockInfo(blockHeight *big.Int) (v BlockAnalyzer, err error) {
	hash, err := d.C.GetBlockHash(blockHeight.Int64())
	if err != nil {
		return nil, err
	}
	blockInfo, err := d.C.GetBlock(hash)
	if err != nil {
		return nil, err
	}
	return &btcBlock{blockInfo, d.Env, blockHeight}, nil
}
func (d *BtcDaemon) CheckConfirmations(confirm uint64) bool {
	return confirm >= d.ConfirmNum
}
func (d *BtcDaemon) CheckUnlock(confirm uint64) bool {
	return confirm >= d.UnlockNum
}

func (d *BtcDaemon) GetTransaction(txId *TxIdInfo, easy bool) (*TxInfo, error) {
	txHash, err := chainhash.NewHashFromStr(txId.TxId)
	if err != nil {
		return nil, err
	}
	txInfo, err := d.C.GetRawTransactionVerbose(txHash)
	if err != nil {
		return nil, err
	}
	txi := &TxInfo{}
	txi.TxId = txId.TxId
	txi.Cnfm = txInfo.Confirmations
	txi.Target = d.ConfirmNum
	txi.ExtValid = true
	txi.BCT = bccore.BC_BTC
	txi.BCS = bccore.STR_BTC
	txi.UnlockNum = d.UnlockNum
	if !easy {

		txi.H = big.NewInt(0).Set(txId.BlockH)
		inSum := big.NewInt(0)
		for _, v := range txInfo.Vin {
			aaIn, err := d.GetFromAddr(v.Txid, v.Vout)
			if err != nil {
				return nil, err
			}
			inSum = inSum.Add(inSum, aaIn.Amt)
			txi.In = append(txi.In, aaIn)
		}
		outSum := big.NewInt(0)
		for _, v := range txInfo.Vout {
			if v.ScriptPubKey.Type == nulldata {
				arr := strings.Split(v.ScriptPubKey.Asm, " ")
				if len(arr) != 2 && arr[0] != opReturnStr {
					continue
				}
				omniRes, err := d.C.OmniGetTransaction(txHash)
				if err == nil && omniRes != nil && omniRes.Propertyid == int64(d.UsdtId) {
					token := strconv.FormatInt(omniRes.Propertyid, 10)
					amt, _ := bccoin.NewCoinAmount(bccore.BC_BTC, bccore.Token(token), omniRes.Amount)
					//omni confirm
					txi.InExt = append(txi.InExt, &AddrAmount{
						Addr: omniRes.Sendingaddress,
						Amt:  amt.Val(),
					})
					txi.OutExt = append(txi.OutExt, &AddrAmount{
						Addr: omniRes.Referenceaddress,
						Amt:  amt.Val(),
					})
					txi.ExtValid = omniRes.Valid
					txi.Token = strconv.FormatInt(omniRes.Propertyid, 10)
					txi.BCS = bccore.STR_USDT
				} else {
					//TODO LOG

				}

			} else if v.ScriptPubKey.Type == pubkeyhash {
				am, _ := bccoin.NewCoinAmountFromFloat(bccore.BC_BTC, "", v.Value)
				out := &AddrAmount{
					Addr: v.ScriptPubKey.Addresses[0],
					Amt:  am.Val(),
				}
				outSum = outSum.Add(outSum, am.Val())
				txi.Out = append(txi.Out, out)
			}
		}
		txi.Fee = inSum.Sub(inSum, outSum)
	}
	return txi, nil
}

func (d *BtcDaemon) GetFromAddr(txId string, vout uint32) (*AddrAmount, error) {
	txHash, err := chainhash.NewHashFromStr(txId)
	if err != nil {
		return nil, err
	}
	tx, err := d.C.GetRawTransactionVerbose(txHash)
	if err != nil {
		return nil, err
	}
	if len(tx.Vout) >= int(vout) {
		aa := &AddrAmount{}
		am1, _ := bccoin.NewCoinAmountFromFloat(bccore.BC_LTC, "", tx.Vout[vout].Value)
		aa.Amt = am1.Val()
		aa.Addr = tx.Vout[vout].ScriptPubKey.Addresses[0]
		return aa, nil
	} else {
		return nil, errors.ERR_TX_OUT_INDEX_OVERFLEW
	}
}

///////////////////////////////////////////////////////////////////////////

type btcBlock struct {
	*wire.MsgBlock
	*chaincfg.Params
	H *big.Int
}

func (b *btcBlock) AnalyzeBlock(addresses []string) (txIds []*TxIdInfo) {
	txIds = make([]*TxIdInfo, 0, 2)
	flag := false
	for _, v := range b.Transactions {
		flag = false
		for _, v1 := range v.TxOut {
			if txscript.IsUnspendable(v1.PkScript) {
				continue
			}
			_, addres, _, err := txscript.ExtractPkScriptAddrs(v1.PkScript, b.Params)
			if err != nil {
				return nil
			}
			if len(addres) == 0 {
				continue
			}
			addr := addres[0].EncodeAddress()
			for _, curAddr := range addresses {
				if curAddr == addr {
					txifo := &TxIdInfo{TxId: v.TxHash().String(), BlockH: b.H}
					txIds = append(txIds, txifo)
					flag = true
					break
				}
			}
			if flag {
				break
			}
		}
	}
	return txIds
}
