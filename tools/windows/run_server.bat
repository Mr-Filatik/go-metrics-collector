go run ..\..\cmd\server\main.go -a=localhost:8080 -i=15 -r=false -d=postgres://user:password@localhost:5432/postgres?sslmode=disable -k=FILATIK_KEY_FOR_HASHING -crypto-key=C:\Projects\Personal\Go\go-metrics-collector\web\crypto\private.pem

pause