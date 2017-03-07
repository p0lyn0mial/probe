# probe
probe is a small tool that will measure your server response times 

# dependencies
for calculation of various percentiles a stats pkg was used, in order to compile the binary go get it.
```sh
go get github.com/montanaflynn/stats
```
note:
the pkg is not vendored at the moment - hopefully the authors of that pkg will not make any changes that could break the API.

at least Go 1.7 is required.

# usage
```sh
make build;make run 
```
by default the app measures response times of the service at ‘https://gitlab.com' over 5 minutes.
TODO: [support for cmd args](https://github.com/p0lyn0mial/probe/blob/master/main.go#L13)

# testing
by default tests are run with --race flag.
```sh
make test 
```

for more verbous output set -v flag.
```sh
make test T_FLAGS="-v"
```
