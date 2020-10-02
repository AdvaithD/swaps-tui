package main

import (
	"context"
	"flag"
	"log"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"swaps-tui/contracts/erc20"
	"swaps-tui/contracts/uniswap/router"
	"swaps-tui/decimal"
	"swaps-tui/mainnet_addrs"
	"unsafe"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var (
	swapExactETHForTokens           = [4]byte{0x7f, 0xf3, 0x6a, 0xb5}
	swapExactTokensForETH           = [4]byte{0x18, 0xcb, 0xaf, 0xe5}
	router_abi, _                   = abi.JSON(strings.NewReader(router.RouterABI))
	method_swapExactETHForTokens, _ = router_abi.MethodById(swapExactETHForTokens[:])
	method_swapExactTokensForETH, _ = router_abi.MethodById(swapExactTokensForETH[:])
	swap_methods                    = map[[4]byte]*abi.Method{
		swapExactETHForTokens: method_swapExactETHForTokens,
		swapExactTokensForETH: method_swapExactTokensForETH,
	}
)

var (
	client_dial = flag.String(
		"client_dial", "ws://192.168.1.2:9551", "could be websocket or IPC",
	)
)

func main() {
	flag.Parse()

	client, err := ethclient.Dial(*client_dial)

	if err != nil {
		log.Fatal(err)
	}

	v := reflect.ValueOf(client).Elem()
	f := v.FieldByName("c")
	rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	concrete_client, _ := rf.Interface().(*rpc.Client)

	something := make(chan common.Hash)

	concrete_client.EthSubscribe(
		context.Background(), something, "newPendingTransactions",
	)

	chainid, _ := client.NetworkID(context.Background())
	signer := types.NewEIP155Signer(chainid)

	type t struct {
		AmountOutMin *big.Int
		Path         []common.Address
		Deadline     *big.Int
		To           common.Address
	}

	type t2 struct {
		AmountIn     *big.Int
		AmountOutMin *big.Int
		Path         []common.Address
		Deadline     *big.Int
		To           common.Address
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	width, height := ui.TerminalDimensions()

	type trading_token struct {
		name  string
		index int
	}

	current_tokens := map[string]*trading_token{}
	trade_data_amount := []float64{}
	labels := []string{}

	current_tokens_2 := map[string]*trading_token{}
	trade_data_amount_2 := []float64{}
	labels_2 := []string{}

	bc := widgets.NewBarChart()
	bc.Data = trade_data_amount
	bc.Labels = labels
	bc.Title = "Uniswap Pending Trades To ETH (aka selling)"
	bc.BarGap = 4
	bc.SetRect(0, 0, width/2, height)
	bc.BarWidth = 2
	bc.BarColors = []ui.Color{ui.ColorRed, ui.ColorGreen, ui.ColorBlue}
	bc.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorBlue)}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorBlack)}

	bc2 := widgets.NewBarChart()
	bc2.Data = trade_data_amount_2
	bc2.Labels = labels_2
	bc2.Title = "Uniswap Pending Trades From ETH (aka buying)"
	bc2.BarGap = 4
	bc2.SetRect(width/2, 0, width, height)
	bc2.BarWidth = 2
	bc2.BarColors = []ui.Color{ui.ColorYellow, ui.ColorWhite, ui.ColorCyan}
	bc2.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorBlue)}
	bc2.NumStyles = []ui.Style{ui.NewStyle(ui.ColorBlack)}

	ui.Render(bc, bc2)

	help_box := widgets.NewParagraph()
	help_box.Title = "keyboard controls"
	help_box.Text = `
e or q exits the program 

r refreshes the data points (not implemented)

press h again to make this help box go away 

tweet at me if you like this - @edgararout
`

	keep_rendering := true
	help_box_shown := false

	go func() {
		for e := range ui.PollEvents() {

			if e.Type == ui.ResizeEvent {
				payload := e.Payload.(ui.Resize)
				width, height := payload.Width, payload.Height
				bc.SetRect(0, 0, width/2, height)
				bc2.SetRect(width/2, 0, width, height)
				ui.Clear()
				ui.Render(bc, bc2)
			}

			if e.Type == ui.KeyboardEvent {
				switch e.ID {
				case "e":
				case "q":
					ui.Close()
					os.Exit(0)
				case "r":

				case "h":
					if help_box_shown {
						help_box_shown = false
						ui.Render(bc, bc2)
						keep_rendering = true
					} else {
						keep_rendering = false
						help_box_shown = true
						help_box.SetRect(width/2, height/2, 0, 0)
						ui.Render(help_box)
					}
				}
			}
		}
	}()

	eighteen := decimal.NewDec(1e18)

	for {
		select {

		case otherwise := <-something:

			if keep_rendering == false {
				continue
			}

			txn, is_pending, _ := client.TransactionByHash(context.Background(), otherwise)

			if is_pending {
				_, _ = signer.Sender(txn)

				if data := txn.Data(); data != nil {

					to := txn.To()

					if to != nil {

						bytecode, _ := client.CodeAt(
							context.Background(), *to, nil,
						)

						isContract := len(bytecode) > 0
						if isContract {

							if *to == mainnet_addrs.UNISWAP_ROUTER {

								if len(data) < 4 {
									continue
								}

								buf := [4]byte{}
								copy(buf[:], data[:4])

								switch buf {
								case swapExactETHForTokens:
									var something t

									if err := method_swapExactETHForTokens.Inputs.Unpack(
										&something, data[4:],
									); err != nil {
										log.Fatal(err)
									}

									dest_addr := something.Path[len(something.Path)-1]
									dest_token, _ := erc20.NewErc20(dest_addr, client)
									dest_token_symbol, _ := dest_token.Symbol(nil)

									if trading, exists := current_tokens_2[dest_token_symbol]; exists {
										trade_data_amount_2[trading.index]++
									} else {
										if len(labels) > 25 {
											continue
										}
										current_tokens_2[dest_token_symbol] = &trading_token{
											name:  dest_token_symbol,
											index: len(trade_data_amount_2),
										}
										labels_2 = append(labels_2, dest_token_symbol)
										val := decimal.NewDecFromBigInt(txn.Value()).Quo(eighteen).String()
										flt, _ := strconv.ParseFloat(val, 64)
										trade_data_amount_2 = append(
											trade_data_amount_2, flt,
										)
									}

									bc2.Data = trade_data_amount_2
									bc2.Labels = labels_2
									ui.Render(bc2)

								case swapExactTokensForETH:
									var something t2

									if err := method_swapExactTokensForETH.Inputs.Unpack(
										&something, data[4:],
									); err != nil {
										log.Fatal(err)
									}

									src_addr := something.Path[0]
									src_token, _ := erc20.NewErc20(src_addr, client)
									src_token_symbol, _ := src_token.Symbol(nil)

									if trading, exists := current_tokens[src_token_symbol]; exists {
										trade_data_amount[trading.index]++
									} else {
										if len(labels) > 25 {
											continue
										}
										current_tokens[src_token_symbol] = &trading_token{
											name:  src_token_symbol,
											index: len(trade_data_amount),
										}
										labels = append(labels, src_token_symbol)
										val := decimal.NewDecFromBigInt(
											something.AmountOutMin,
										).Quo(eighteen).String()
										flt, _ := strconv.ParseFloat(val, 64)
										trade_data_amount = append(
											trade_data_amount, flt,
										)
									}

									bc.Data = trade_data_amount
									bc.Labels = labels
									ui.Render(bc)
								}
							}
						}
					}
				}
			}
		}
	}
}
