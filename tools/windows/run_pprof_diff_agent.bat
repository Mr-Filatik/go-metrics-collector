cd ..\..\
go tool pprof -http=":9090" -diff_base=profiles/agent/base.pprof profiles/agent/result.pprof

pause