import jwt
from jwt import PyJWKClient

def decode_gitlab_jwt(token: str):
    try:
        instance = input("Input: ")
        jwks_url = f"{instance}/oauth/discovery/keys"
        
        jwk_client = PyJWKClient(jwks_url)
        signing_key = jwk_client.get_signing_key_from_jwt(token)

        decoded = jwt.decode(
            token,
            signing_key.key,
            algorithms=["RS256"],
            audience=None,   # set if needed
            options={"verify_aud": False}
        )

        return decoded

    except jwt.ExpiredSignatureError:
        print("Token has expired")
    except jwt.InvalidTokenError as e:
        print(f"Invalid token: {e}")

    return None


if __name__ == "__main__":
    token = input("token: ")
    result = decode_gitlab_jwt(token)
    if result:
        print("Decoded JWT:")
        print(result)
