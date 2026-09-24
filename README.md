# FedRAMPHound

BloodHound OpenGraph collector for the FedRAMP marketplace.

FedRAMPHound takes JSON data exported from the [FedRAMP Marketplace](https://www.fedramp.gov/marketplace/products/) and converts it into a BloodHound OpenGraph payload. Once imported, you can explore products, the federal agencies that use them, and the assessors (3PAOs) that evaluated them, all as a connected graph instead of a flat list.

## Requirements

- Go 1.21 or later (to build from source)
- A running BloodHound instance with OpenGraph support

## Downloading the Marketplace Data

FedRAMPHound works from JSON exports of the FedRAMP Marketplace rather than scraping the site directly. To get the data:

1. Go to the [FedRAMP Marketplace](https://www.fedramp.gov/marketplace/products/).
2. Use the Products, Agencies, and Assessors tabs/filters to narrow down the results you want (or leave filters open to export everything).
3. Use the marketplace's export option to export each view as a JSON data file.
4. Save the three files locally, for example:
   - `products.json`
   - `agencies.json`
   - `assessors.json`

## Running the Tool

Build the binary from source:

```bash
git clone https://github.com/lum8rjack/FedRAMPHound.git
cd FedRAMPHound
go build -o fedramphound .
```

### Collect: Build the OpenGraph JSON

Point the `collect` command at the three JSON exports to build the OpenGraph payload:

```bash
./fedramphound collect \
  -agencies agencies.json \
  -assessors assessors.json \
  -products products.json \
  -output fedramp-opengraph.json
```

This produces `fedramp-opengraph.json`, an OpenGraph-formatted file containing nodes for products, agencies, and assessors, along with the edges connecting them (which agencies use which products, which assessors evaluated which products, and so on).

### Importing into BloodHound

Once you have `fedramp-opengraph.json`, you can bring it into BloodHound in one of two ways:

**Option 1: Upload through the BloodHound UI**
Open BloodHound, go to the file upload / ingest option, and upload `fedramp-opengraph.json` directly.

**Option 2: Upload via the `upload` command**
FedRAMPHound also includes an `upload` command that pushes the extension schema (`fedramp-marketplace-schema.json`) to a BloodHound instance over the API:

```bash
./fedramphound upload \
  -endpoint https://bloodhound.example.com \
  -schema fedramp-marketplace-schema.json \
  -token-id YOUR_TOKEN_ID \
  -token-key YOUR_TOKEN_KEY
```

You can authenticate with either a token ID/key pair or a bearer JWT (`-bearer`). Once the schema and graph data are loaded, open BloodHound and start exploring the FedRAMP marketplace as a graph.

