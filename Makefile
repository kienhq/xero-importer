generate_access_token:
	python3 xero.py

generate_accounts:
	ACTION=generate_accounts go run .

upload_accounts:
	ACTION=upload_accounts go run .