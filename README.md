# Steampipe plugin for Blockchain.com

This is a [Steampipe](https://steampipe.io) plugin that interfaces with the [Blockchain.com](https://blockchain.com) API and returns information about Bitcoin wallets and their transactions.

## Usage

- [ ] For each table that you want to expose via SQL:
    - [ ] Copy the `samplerest/table_samplerest_one_model.go` file
    - [ ] Rename it to describe the service (instead of `samplerest`) and the entity listed (instead of `one_model`). For example, `table_github_repository`
    - [ ] Change the `Name` and `Description`
    - [ ] If your model doesn't support searching to return a subset of items, delete the `List.KeyColumns` field, and any other places marked with `Delete if your API doesn't suport searching over all instances`
    - [ ] Add/edit all column names types and descriptions in `Columns` to match whatever is exposed by the API. The `Name` field will be seen by SQL, and the `Transform` field is used to match the objects that are returned by the `List` and `Get` functions
    - [ ] Rename the `OneModel` struct, and edit it to match the data exposed by the API. The field names should match with the names passed to the `Columns.Transform` configs above
    - [ ] Edit the `listOneModel` function to contact the API and get the results. You have available the `config` var, which holds API credentials, and possibly the `realQueryString` and/or `realQueryJson` variables, for filtering
    - [ ] Complete the `listOneModel` function to make it return all data returned by the API
    - [ ] Edit the `getOneModel` function to contact the API and get a single result. You have available the `config` var, which holds API credentials, and the `id` var, which holds the ID of the single object
    - [ ] Complete the `getOneModel` function to make it return the data of a single item
    - [ ] Rename the `listOneModel` and `getOneModel` functions to something that matches the actual objects. For example, `listRepository` and `getRepository` for the file `table_github_repository.go`

## Testing

Run `make`, then run `steampipe query`. Run `.inspect` inside of it to ensure that the plugin is loaded.

Alternatively, run `go build -o ~/.steampipe/plugins/hub.steampipe.io/plugins/jreyesr/blockchain@latest/steampipe-plugin-blockchain.plugin *.go`.