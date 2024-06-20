# Generate & Upload accounts to XERO

## Requirements
- Python3
- Golang 1.20
- Docker
- docker-compose

## Configuration
- Create .env file from .env.sample
- Access to your xero developer dashboard here https://developer.xero.com/app/manage
- Create a new app and import the following variables to env file

```
CLIENT_ID=""
CLIENT_SECRET=""
```

## Setup Docker

### 1. Run docker container

```shell
docker-compose up -d
```

### 2. Access to container

```shell
docker exec -it xero-importer /bin/sh
```

### 3. Install go & python dependencies

```shell
go mod tidy && pip3 install -r requirements.txt
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
- To change the number of files, initial code number or number of COA for each file, you can modify these variables:
```
INIT_COA_NUMBER=30000
NUM_GENERATED_COA=500
NUM_GENERATED_FILES=2
```

### 3. upload_accounts

- Upload accounts will read files from FILES variable to upload COA to XERO

```shell
make upload_accounts
```
