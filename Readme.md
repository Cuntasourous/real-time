# Forum

## Description
This project is a web-based forum application designed to facilitate communication and interaction among users. It allows users to create and categorize posts, comment on posts, and like or dislike posts and comments. The forum includes features such as user authentication, session management, and post filtering. The backend is powered by SQLite for data storage, and the application is containerized using Docker for easy deployment. This project aims to provide a comprehensive understanding of web development, database management, and containerization.


## Setup and Installation
```sh
git clone https://learn.reboot01.com/git/araed/forum
cd forum 
```
To run the project
```go
go run .
```

## Docker 
To run docker file do the following or simply run sh docker.sh
```
docker image build -f Dockerfile -t dockerize .

docker container run -p 8080:8080 --name forum dockerize
```

## Contributors 
- Maram Sagheer (msagheer)
- Ameena Raed (araed)
- Zahraa Fadhel (zfadhel)
- Omar Albinkhalil (ok)
- Sara Abdulla (sabdulla)