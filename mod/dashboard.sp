dashboard "single_wallet_monitor_dashboard" {
  title = "Single Wallet Monitor Dashboard"

  container {
    input "wallet" {
      title = "Wallet ID:"
      width = 4
      type  = "combo"
      placeholder = "enter a wallet id in the base58 format"

      sql = <<-EOQ
        with wallets as (select unnest($1::text[]) as id)
        select id as value, id as label from wallets; 
      EOQ

      args = [var.monitored_wallets]
    }
  }
  
  container {
    card {
      sql = <<EOQ
        select
          case when final_balance = 0 then 'alert' else 'ok' end as type,
          'Current Balance' as label,
          final_balance as value,
          address
        from blockchain_wallet
        where address = $1;
        EOQ
    
      icon  = "currency_bitcoin"
      width = 4
      href  = "https://www.blockchain.com/explorer/addresses/btc/{{.address}}"
      args =  [self.input.wallet.value]
    }

    card {
      sql = <<EOQ
        select
          total_received as value,
          'Total Received' as label
        from blockchain_wallet
        where address = $1;
        EOQ
    
      icon  = "move_to_inbox"
      width = 4
      args =  [self.input.wallet.value]
    }

    card {
      sql = <<EOQ
        select
          total_sent as value,
          'Total Sent' as label
        from blockchain_wallet
        where address = $1;
        EOQ
    
      icon  = "outbox"
      width = 4
      args =  [self.input.wallet.value]
    }
  }
  
  container {
    title = "Transactions for wallet"
    
    table {
      sql = <<EOQ
        select
          'Click here' as "Link",
          hash as "Hash",
          inputs_value as "Inputs", outputs_value as "Outputs", fee as "Fee",
          time as "Time",
          wallet_balance as "Wallet Involvement"
        from blockchain_transaction
        where wallet = $1
        order by time desc limit 10;
        EOQ

      width = 12
      args =  [self.input.wallet.value]

      column "Hash" {
        display = "none"
      }
      column "Link" {
        href = "https://www.blockchain.com/explorer/transactions/btc/{{.'Hash'}}"
      }
    }
  }

  container {
    chart {
      sql = <<EOQ
        select
          to_char(date_trunc('day', time::date), 'YYYY-MM-DD'),
          sum(case when wallet_balance > 0 then wallet_balance else 0 end) / 1e8 as "Received",
          sum(case when wallet_balance < 0 then wallet_balance else 0 end) / 1e8 as "Sent"
        from blockchain_transaction
        where wallet = $1
        group by 1;
      EOQ

      width = 12
      args =  [self.input.wallet.value]

      axes {
        x {
          type = "time"
        }
      }
      series "Received" {
        color = "green"
      }
      series "Sent" {
        color = "red"
      }
    }
  }
}