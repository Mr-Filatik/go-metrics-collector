cd ..\..\
go tool pprof -http=":9090" -seconds=30 http://localhost:8081/debug/pprof/heap

pause