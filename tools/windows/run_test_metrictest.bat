echo off

cd ..\..\
cd cmd\server\
echo Run build for server application
go build -o server.exe main.go
echo Build for server application is done

cd ..\..\
cd cmd\agent\
echo Run build for server application
go build -o agent.exe main.go
echo Build for agent application is done

cd ..\..\

echo Run TestIteration1
metricstest-windows-amd64 -test.run=^TestIteration1$ -binary-path=cmd/server/server
echo ====================================================================================================

echo Run TestIteration2
metricstest-windows-amd64 -test.v -test.run=^TestIteration2$ -binary-path=cmd/server/server -source-path=. -agent-binary-path=cmd/agent/agent -server-port 8080
echo ====================================================================================================

echo Run TestIteration3
metricstest-windows-amd64 -test.v -test.run=^TestIteration3$ -binary-path=cmd/server/server -source-path=. -agent-binary-path=cmd/agent/agent -server-port 8080
echo ====================================================================================================

echo Run TestIteration4
metricstest-windows-amd64 -test.v -test.run=^TestIteration4$ -binary-path=cmd/server/server -source-path=. -agent-binary-path=cmd/agent/agent -server-port 8080
echo ====================================================================================================

echo Run TestIteration5
metricstest-windows-amd64 -test.v -test.run=^TestIteration5$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -source-path=.
echo ====================================================================================================

echo Run TestIteration6
metricstest-windows-amd64 -test.v -test.run=^TestIteration6$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=$SERVER_PORT -source-path=.
echo ====================================================================================================

echo Run TestIteration7
metricstest-windows-amd64 -test.v -test.run=^TestIteration7$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=$SERVER_PORT -source-path=.
echo ====================================================================================================

echo Run TestIteration8
metricstest-windows-amd64 -test.v -test.run=^TestIteration8$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8080 -source-path=.
echo ====================================================================================================

echo Run TestIteration9
metricstest-windows-amd64 -test.v -test.run=^TestIteration9$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -file-storage-path=../../temp_metrics_test.json -server-port=8080 -source-path=.
echo ====================================================================================================

echo Run TestIteration10
metricstest-windows-amd64 -test.v -test.run=^TestIteration10[AB]$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -database-dsn='postgres://user:password@localhost:5432/postgres?sslmode=disable' -server-port=8080 -source-path=.
echo ====================================================================================================

pause

echo on