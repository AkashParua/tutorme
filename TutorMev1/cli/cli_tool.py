import requests
import jwt

base_url = "http://localhost:8080"
secret_key = "my_secret_key"

def login(username, password):
    url = base_url + "/login"
    data = {
        "username": username,
        "password": password
    }
    response = requests.post(url, json=data)
    if response.status_code == 200:
        return response.json()["token"]
    else:
        return None

def get_questions(token):
    url = base_url + "/questions"
    headers = {
        "Authorization": "Bearer " + token
    }
    response = requests.get(url, headers=headers)
    if response.status_code == 200:
        return response.json()
    else:
        return None

def verify_token(token):
    try:
        jwt.decode(token, secret_key, algorithms=["HS256"], issuer="my_issuer")
        return True
    except jwt.exceptions.InvalidTokenError:
        return False

def main():
    username = input("Enter username: ")
    password = input("Enter password: ")

    token = login(username, password)
    if token is None:
        print("Login failed")
        return

    if not verify_token(token):
        print("Token verification failed")
        return

    questions = get_questions(token)
    if questions is None:
        print("Failed to get questions")
        return

    print(questions)

if __name__ == "__main__":
    main()