# Albion Prices

Albion Prices is a tool for importing prices for Albion Online video game provided by the [Albion Online Data Project](https://www.albion-online-data.com/)'s API into a Google Sheets spreadsheet

## Requirements

* [Go](https://golang.org/dl/)
* Google Sheets API for Go
  ```shell
  go get -u google.golang.org/api/sheets/v4 golang.org/x/oauth2/google
  ```

## Installation

1. Download the repo
2. Follow steps 1 and 2 from [Sheets API quickstart](https://developers.google.com/sheets/api/quickstart/go)

   Alternatively, you can reuse a different Cloud Platform project with Google Sheets API enabled
3. Move credentials.json to this project's root directory
4. Create a spitesheet in GoogleSheets where you'd wish for data to end up
5. Get the spreadsheet ID, sheet name and sheet location to insert your data into. ([If you need help](https://developers.google.com/sheets/api/guides/concepts))
6. Open `albion_prices.go` in your text editor of choice
7. In lines

   ```go
   const spreadsheetID string = ""
   const sheetName string = "MarketData"
   const sheetPosition string = "A1"
   ```

   Place the spreadsheet ID, sheet name and sheet location from step 5 between their respective quotes
8. Open your terminal of choice, navigate to Albion Price's root directory
9. Run Albion Prices in the terminal for the first time

   ```shell
   go run albion_prices.go
   ```
10. You will be given a link for authorizing access to your Google Sheets, open it in the browser
11. Log into your account to authorize access to Sheets API
   
    If you see "That app isn't verified!" screen, click to show advanced, then click "Go to Quickstart(unsafe)". If in step 2 you used a different Cloud Platform project then the name Quickstart will be the name of that project instead. 
    
    Note: This is completely safe, each user creates their own Google project, with access only to their own sheets.
12. Copy the code you're given and paste it the terminal, hit enter
13. After successful authorization the program will run for the first time. When it finishes, check if the data was written into your sheet.

## Usage

If you've followed the installation steps successfully, then after the first authorization the subsequent runs don't require authorization and can be run at-will with `go run albion_prices.go`

You can compile the program into an executable for easier use:
```shell
go build albion_prices.go
```
Note that if you do, any changes to `albion_prices.go` won't have any effect on the executable.

The data generated into the sheet will be in a form of table, one item per row, one city per two columns. The number in an item's row and city's left column is the minimum sell price of said item in said city, the number in the city's right colum is the item's maximum buy price in said city. Zeroes or blank cells represent either no sell/buy orders or no data avaiable. First row contains the names of the cities, the first column contains the IDs of the items.

Item IDs are avaiable in [Albion Online's binary data dump](https://github.com/broderickhyman/ao-bin-dumps/blob/master/formatted/items.txt)

### Querying different or additional items

Each line in `enchantableResources.txt`, `enchantableItems.txt` and `unenchantableItems.txt` contains data about which items to query for prices from Albion Online Data Project:

* `unenchantableItems.txt` contains item IDs which will be queried as-is
* `enchantableItems.txt` contains item IDs which will be queried as-is, and will also query enchant tiers for said items
* `enchantableResources.txt` contains resources and materials will be queried as-is, and will also query enchant tiers for said resources or materials

Enchantable items and resources are split because their enchantment tier is denoted differently in item IDs