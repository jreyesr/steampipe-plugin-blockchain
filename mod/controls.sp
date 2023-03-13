variable "monitored_wallets" {
    type = list(string)
    description = "A list of wallets to monitor"   
}

control "nonempty_wallet" {
    title = "Wallet must have funds"
    sql = <<EOT
        select
            address as resource,
            case when final_balance = 0 then 'alarm' else 'ok' end as status,
            case when final_balance = 0 then format('Wallet %s is empty', address) else format('Wallet %s has funds', address) end as reason,
            total_received, total_sent,
            final_balance
        from blockchain_wallet
        where address = any($1);
        EOT
    param "addresses" {
        default = var.monitored_wallets
    }
}

control "recently_alive_wallet" {
    title = "Wallet must have made a transaction recently"
    sql = <<EOT
        with data as (
            select
                hash,
                wallet,
                inputs_value as amount,
                time,
                EXTRACT(EPOCH FROM(now() - time)) as timedelta,
                EXTRACT(EPOCH FROM(now() - time)) <= 31536000 as tx_in_last_year
            from blockchain_transaction
            where wallet = any($1)
            order by time desc
            limit 1
        ) select
            hash as resource,
            case when not tx_in_last_year then 'alarm' else 'ok' end as status,
            case when not tx_in_last_year then format('Wallet %s has not made a transaction in a year', wallet) else format('Wallet %s has made at least a transaction in the past year', wallet) end as reason,
            amount, to_char(time, 'YYYY-MM-DD') as last_tx
        from data;
        EOT
    param "addresses" {
        default = var.monitored_wallets
    }
}