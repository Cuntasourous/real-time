FROM golang:latest
LABEL contributers="Ameena Raed Maram Sagheer, Zahraa Fadhel, Omar Albinkhalil, Sara Abdulla"
LABEL project="Forum"
#sets the working directory, all commands will be run in the /app directory
WORKDIR /app
# Copy the local package files to the container's workspace
COPY . .
#RUN: execute commands inside the container during the image build process
# Downloads the modules specified in go.mod file
RUN go mod download
#go build: This is a command that tells Go to turn your Go code into a ready-to-run program. 
# -o flag is used to specify the output file name
RUN go build -o main
EXPOSE 8080
CMD ["./main"]
