benchmark "blockchain_audit" {
    title = "Blockchain Audit"
    children = [
        benchmark.btc_wallets,
        benchmark.btc_txs,
    ]
}

benchmark "btc_wallets" {
    title = "1. Wallets"
    children = [
        control.nonempty_wallet,
    ]
}

benchmark "btc_txs" {
    title = "2. Transactions"
    children = [
        control.recently_alive_wallet,
    ]
}