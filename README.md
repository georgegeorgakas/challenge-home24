# challenge-home24

Steps to execute.

1. go get (if you are missing any packages)
2. go run main.go
3. curl --location --request POST 'http://localhost:8080/event' \
   --header 'Content-Type: application/json' \
   --data-raw '{
      "url" : "https://home24.career.softgarden.de/en//"
   }'
4. Run the curl above either from terminal or postman to receive the website of the data you want.

When you receive your response you can send another curl with different URL.
