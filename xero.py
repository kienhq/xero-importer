from requests.auth import HTTPBasicAuth
from requests_oauthlib import OAuth2Session
import os
from dotenv import load_dotenv, set_key

load_dotenv()

# Load Xero API credentials from env
CLIENT_ID = os.getenv('CLIENT_ID')
CLIENT_SECRET = os.getenv('CLIENT_SECRET')
REDIRECT_URI = os.getenv('REDIRECT_URI')
AUTHORIZATION_BASE_URL = os.getenv('AUTHORIZATION_BASE_URL')
TOKEN_URL = os.getenv('TOKEN_URL')
SCOPE = os.getenv('SCOPE').split(',')


def get_access_token():
    # Step 1: Get authorization URL
    oauth = OAuth2Session(CLIENT_ID, redirect_uri=REDIRECT_URI, scope=SCOPE)
    authorization_url, state = oauth.authorization_url(AUTHORIZATION_BASE_URL)

    print(f'Please go to {authorization_url} and authorize access.')

    # Step 2: Get the authorization code from the callback URL
    redirect_response = input('Paste the full redirect URL here: ')

    # Step 3: Fetch the access token
    oauth.fetch_token(TOKEN_URL, authorization_response=redirect_response, auth=HTTPBasicAuth(CLIENT_ID, CLIENT_SECRET))

    access_token = oauth.token['access_token']
    resp = oauth.get('https://api.xero.com/connections')
    if resp.status_code == 200:
        return access_token, resp.json()[0].get('tenantId')
    raise Exception(resp.text)


if __name__ == '__main__':
    access_token, tenant_id = get_access_token()
    print(f"access_token: {access_token} - tenant_id: {tenant_id}")
    set_key(dotenv_path='.env', key_to_set='ACCESS_TOKEN', value_to_set=access_token)
    set_key(dotenv_path='.env', key_to_set='TENANT_ID', value_to_set=tenant_id)
