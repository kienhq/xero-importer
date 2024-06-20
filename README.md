# Generate & Upload accounts to XERO

## Requirements
- Python3
- Golang 1.20

## Setup

```shell
go mod tidy && pip3 install -r requirements.txt
```

## Configuration
- Access to your xero developer dashboard here https://developer.xero.com/app/manage
- Create a new app and import the following variables to env file

```
CLIENT_ID=""
CLIENT_SECRET=""
```

## Run

### 1. generate access_token

```shell
make generate_access_token
```

### 2. generate_accounts

```shell
make generate_accounts
```

### 3. upload_accounts

```shell
make upload_accounts
```
