# Test rest api
curl -H "Content-Type: application/json" -X POST -d '{"username": "user1", "password":"123456" }' "http://127.0.0.1:8081/v1/auth/login"