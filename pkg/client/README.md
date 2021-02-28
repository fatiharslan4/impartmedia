## Impart Wealth Backend Client

This is a CLI application for managing the impart wealth backend. 

## Building

``` bash
go build -mod=vendor -o impartcli cmd/client/main.go && \
ln -sf $(pwd)/impartcli /usr/local/bin/impart && \
impart -h
```

## Usage

All help is context sensitive, so if you run 
``` bash
impart --help # (or -h)
``` 
it will show the required global variables, as well as the available commands.  

Each command also has a context sensitive help menu, so you can run a commands help and see the 
documentation for creating a user.
```bash
impart create-user --help
```
