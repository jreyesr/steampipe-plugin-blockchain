# Steampipe plugin for BTC.com

This is a [Steampipe](https://steampipe.io) plugin that interfaces with the [BTC.com](https://btc.com) API and returns information about Bitcoin wallets and their transactions.

## Configuration

No configuration is required to use this plugin. Copy the `config/blockchain.spc` file to the Steampipe `config` directory.

## Usage

> **NOTE:** This examples use the wallet `1MusKqjbk497v4Jf1bkgSpKb4aUhjzfoqA`. This wallet was found on the [Bitcoin Abuse Database](https://www.bitcoinabuse.com), so it may be involved in shady operations! Indeed, it's reported multiple times as a "cryptocurrency giveaway scam".  
Further muddying the waters, the reports themselves look like Platinum A+ Certified Spam (TM) too, advertising "bitcoin recovery services", soooo...  
In other words, treat this wallet address with caution. It may be completely innocent, or it may be evil. Don't just go around sending it money because it appeared on these examples.  
The report is at <https://www.bitcoinabuse.com/reports/1MusKqjbk497v4Jf1bkgSpKb4aUhjzfoqA>

List details about a wallet:

```sql
select * from blockchain_wallet where address='1MusKqjbk497v4Jf1bkgSpKb4aUhjzfoqA'
```

List all transactions that involve a certain account:

```sql
select * from blockchain_transaction where wallet='1MusKqjbk497v4Jf1bkgSpKb4aUhjzfoqA' order by time desc
```

List details for a single transaction:

```sql
select * from blockchain_transaction where hash='c15459fc73e0d6c647cddc003beab6241475c479ed45dc7ae3743164f5cbd100'
```

## Testing

Run `make`, then run `steampipe query`. Run `.inspect` inside of it to ensure that the plugin is loaded.

Alternatively, run `go build -o ~/.steampipe/plugins/hub.steampipe.io/plugins/jreyesr/blockchain@latest/steampipe-plugin-blockchain.plugin *.go`.