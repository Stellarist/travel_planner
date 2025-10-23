# Start Travel Planner - Backend and Frontend

Write-Host "Starting Travel Planner..." -ForegroundColor Green

# Start backend in a new PowerShell window
Write-Host "Starting Backend (Go)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd c:\Data\repos\travel_planner\backend; go run main.go"

# Wait a moment before starting frontend
Start-Sleep -Seconds 2

# Start frontend in a new PowerShell window
Write-Host "Starting Frontend (npm)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd c:\Data\repos\travel_planner\frontend; npm run dev"

Write-Host "`nBoth processes started in separate windows!" -ForegroundColor Green
Write-Host "Close those windows to stop the services." -ForegroundColor Yellow
