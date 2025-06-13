cd ..\..\

start "" godoc -http=:6060

timeout /t 3 >nul

start "" "C:\Program Files\Google\Chrome\Application\chrome.exe" "http://localhost:6060/pkg/github.com/Mr-Filatik/go-metrics-collector/internal/?m=all"