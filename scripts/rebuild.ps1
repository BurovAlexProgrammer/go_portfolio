# Имя контейнера

$containerName = "go_portfolio_container"

# Останавливаем и удаляем предыдущий контейнер
Write-Host "Stopping and removing old container (if exists)..."
docker stop $containerName 2>$null
docker rm $containerName 2>$null

# Пересобираем образ
Write-Host "Rebuilding Docker image..."
docker build -t go_portfolio .

# Запускаем контейнер с привязкой портов
Write-Host "Starting new container..."
docker run -p 8083:8080 --name $containerName go_portfolio