# Generate & Upload accounts to XERO

## Requirements
- Python3
- Golang 1.20

## Setup

```shell
go mod tidy && pip3 install -r requirements.txt
```

## Configuration
- Create .env file from .env.sample
- Access to your xero developer dashboard here https://developer.xero.com/app/manage
- Create a new app and import the following variables to env file

```
CLIENT_ID=""
CLIENT_SECRET=""
```

## Run

### 1. generate access_token

- Run this command firstly to generate access_token and tenant_id in .env file

```shell
make generate_access_token
```

### 2. generate_accounts

- Config the COA_PATH to specify the directory you want to save the coa files
- After running, list of files will be saved into FILES variable within .env file

```shell
make generate_accounts
```

### 3. upload_accounts

- Upload accounts will read files from FILES variable to upload COA to XERO

```shell
make upload_accounts
```
