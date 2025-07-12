//Docker build
docker build -t go_portfolio .

//Docker run
docker run -p 8083:8080 go_portfolio

//Build executable
go build -o ./bins/goTgPortfolio.exe ./cmd

                    